/*
   Copyright The Accelerated Container Image Authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package cache

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/data-accelerator/dadi-p2proxy/pkg/p2p/client"
	"github.com/data-accelerator/dadi-p2proxy/pkg/p2p/metagc"
	"github.com/data-accelerator/dadi-p2proxy/pkg/p2p/util"
	"github.com/dgraph-io/ristretto"
	log "github.com/sirupsen/logrus"
)

// Config set cache size, entry num and cache media(fs)
type Config struct {
	CacheSize  int64
	MaxEntry   int64
	CacheMedia string
}

// FileCachePool provides basic interface for cache access
type FileCachePool interface {
	// GetLen fetch value length if hit
	GetLen(path string) (int64, bool)
	// PutLen set value length
	PutLen(path string, length int64) bool
	// GetOrRefill try to fetch cache value, call `fetch` if not hit
	GetOrRefill(path string, offset int64, count int, fetch func() ([]byte, error)) ([]byte, error)
	// GetHost get P2P Host for key
	GetHost(path string) (string, bool)
	// PutHost store P2P Host for key
	PutHost(path string, host string) bool
	// DelHost clear P2P Host for key
	DelHost(path string)
}

type fileCachePoolImpl struct {
	fileCache *ristretto.Cache
	memCache  *ristretto.Cache
	media     string
	funcMeta  map[string]map[string]int64
	funcLock  sync.Mutex
	lock      sync.Mutex
}

func (c *fileCachePoolImpl) GetOrRefill(path string, offset int64, count int, fetch func() ([]byte, error)) ([]byte, error) {
	key := filepath.Join(c.media, path, strconv.FormatInt(offset, 10))
retry:
	c.lock.Lock()
	c.fileCache.Wait()
	val, found := c.fileCache.Get(key)
	if !found {
		var err error
		if val, err = newFileCacheItem(key, count); err != nil {
			return nil, err
		}
		c.fileCache.Set(key, val, 0)
		c.PutFuncMeta(path, offset)
	}
	c.lock.Unlock()
	item := val.(*fileCacheItem)
	item.lock.Lock()
	if item.file == nil {
		item.lock.Unlock()
		log.Warnf("File %s already drop, retry!", key)
		goto retry
	}
	defer item.lock.Unlock()
	value := item.Val()
	if len(value) == 0 {
		if err := item.Fill(fetch); err != nil {
			return nil, err
		}
		value = item.Val()
	}
	return value, nil
}

func (c *fileCachePoolImpl) GetLen(path string) (int64, bool) {
	key := filepath.Join(path, "metainfo")
	c.memCache.Wait()
	val, found := c.memCache.Get(key)
	if !found {
		return 0, false
	}
	return val.(int64), true
}

func (c *fileCachePoolImpl) PutLen(path string, len int64) bool {
	key := filepath.Join(path, "metainfo")
	c.memCache.Set(key, len, 1)
	return true
}

func (c *fileCachePoolImpl) PutFuncMeta(path string, offset int64) bool {
	funcName := util.GetMetaKey(path)
	if funcName == "" {
		return true
	}
	c.funcLock.Lock()
	defer c.funcLock.Unlock()

	key := filepath.Join(path, strconv.FormatInt(offset, 10))
	if set, ok := c.funcMeta[funcName]; ok {
		if _, ok = set[key]; ok {
			return true
		}
		set[key] = offset
		return true
	}
	set := make(map[string]int64)
	set[key] = offset
	c.funcMeta[funcName] = set
	return true
}

func (c *fileCachePoolImpl) DelFuncMeta(key string) {
	path := key[strings.Index(key, "/")+1:]
	funcName := util.GetMetaKey(path)
	if funcName == "" {
		return
	}

	c.funcLock.Lock()
	set, ok := c.funcMeta[funcName]
	if !ok { //已经被gc了
		return
	}

	_, ok = set[path]
	if !ok { // path doesn't exist.
		return
	}

	// delete all metas of function-name
	delete(set, path)
	for k, _ := range set {
		key := filepath.Join(c.media, k)
		c.fileCache.Del(key)
		delete(set, k)
	}
	delete(c.funcMeta, funcName)
	c.funcLock.Unlock()

	// send gcRPC
	hostIp, err := client.GetOutBoundIP()
	if err != nil {
		log.Error(err)
	}
	req := metagc.MetaGcRequest{ActionName: funcName, Kind: funcName, InvokerIp: hostIp}
	ctx := context.Background()
	client := client.GetGRPCClient()
	if _, err := client.GcMetadata(ctx, &req); err != nil {
		client.GcMetadata(ctx, &req)
	}
}

func (c *fileCachePoolImpl) GetHost(path string) (string, bool) {
	key := filepath.Join(path, "upstream")
	c.memCache.Wait()
	val, found := c.memCache.Get(key)
	if !found {
		return "", false
	}
	return val.(string), found
}

func (c *fileCachePoolImpl) PutHost(path string, host string) bool {
	key := filepath.Join(path, "upstream")
	c.memCache.Set(key, host, 1)
	return true
}

func (c *fileCachePoolImpl) DelHost(path string) {
	key := filepath.Join(path, "upstream")
	c.memCache.Del(key)
}

// NewCachePool creator for FileCachePool
func NewCachePool(config *Config) FileCachePool {
	atomic.StoreInt32(&fdCnt, 0)
	if err := os.MkdirAll(config.CacheMedia, 0755); err != nil {
		log.Fatalf("Mkdir %s fail! %s", config.CacheMedia, err)
	}
	cachePool := &fileCachePoolImpl{}
	cachePool.funcMeta = make(map[string]map[string]int64)
	var err error
	if cachePool.fileCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     config.CacheSize,
		BufferItems: 64,
		OnExit: func(val interface{}) {
			item := val.(*fileCacheItem)
			cachePool.DelFuncMeta(item.key)
			item.Drop()
		},
		Cost: func(val interface{}) int64 {
			item := val.(*fileCacheItem)
			item.lock.RLock()
			defer item.lock.RUnlock()
			info, err := item.file.Stat()
			if err != nil {
				return 0
			}
			return info.Size()
		},
	}); err != nil {
		panic(err)
	}
	if cachePool.memCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     config.MaxEntry,
		BufferItems: 64,
	}); err != nil {
		panic(err)
	}
	cachePool.media = config.CacheMedia
	return cachePool
}
