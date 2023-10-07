package server

import (
	"math/rand"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/data-accelerator/dadi-p2proxy/pkg/p2p/client"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
)

func DispatcherHandler(w http.ResponseWriter, req *http.Request) {
	const blobsPattern = "/v2/.*/blobs/.*"
	if matched, _ := regexp.MatchString(blobsPattern, req.URL.Path); matched {
		ServerBlob(w, req)
	} else if req.Method == http.MethodHead {
		ResolvedManifest(w, req)
	} else {
		GetManifest(w, req)
	}
}

func urlPathExtract(path string) (string, string) {
	ss := strings.Split(path, "/")
	return ss[2] + "/" + ss[3], ss[5]
}

func ResolvedManifest(w http.ResponseWriter, r *http.Request) {
	var (
		mediaType string
		size      int64
		dgst      digest.Digest
		fi        fs.FileInfo
		err       error
	)
	name, ref := urlPathExtract(r.URL.Path)
	retry := client.RETRY
	for retry > 0 {
		if mediaType, dgst, err = client.GetManifestInfoByTag(name, ref); mediaType != "" {
			if fi, err = os.Stat(filepath.Join(client.ContainerdRoot, "blobs/sha256", dgst.String()[7:])); err == nil {
				size = fi.Size()
				break
			}
		}
		retry--
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	}
	if err != nil { //using registry
		w.WriteHeader(http.StatusNotFound)
		return
	}
	logrus.Info("ResponseWriter@Content-length: ", size)
	w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
	w.Header().Set("Content-Type", mediaType)
	w.Header().Set("Content-Length", fmt.Sprint(size))
	w.Header().Set("Docker-Content-Digest", dgst.String())
	w.Header().Set("Etag", fmt.Sprintf(`"%s"`, dgst))
}

func GetManifest(w http.ResponseWriter, r *http.Request) {
	var (
		p   []byte
		err error
	)
	_, ref := urlPathExtract(r.URL.Path)
	retry := client.RETRY
	for retry > 0 {
		if p, err = os.ReadFile(filepath.Join(client.ContainerdRoot, "blobs/sha256", ref[7:])); err == nil {
			break
		}
		retry--
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	}
	if err != nil { //using registry
		w.WriteHeader(http.StatusNotFound)
		return
	}
	logrus.Info("ResponseWriter@Content-length: ", len(p))
	w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
	w.Header().Set("Content-Type", r.Header["Accept"][0])
	w.Header().Set("Content-Length", fmt.Sprint(len(p)))
	w.Header().Set("Docker-Content-Digest", ref)
	w.Header().Set("Etag", fmt.Sprintf(`"%s"`, ref))
	w.Write(p)
}

func ServerBlob(w http.ResponseWriter, r *http.Request) {
	var (
		f    *os.File
		size int64
		err  error
	)
	_, ref := urlPathExtract(r.URL.Path)
	retry := client.RETRY
	for retry > 0 {
		if fi, err := os.Stat(filepath.Join(client.ContainerdRoot, "blobs/sha256", ref[7:])); err == nil {
			size = fi.Size()
			if f, err = os.Open(filepath.Join(client.ContainerdRoot, "blobs/sha256", ref[7:])); err == nil {
				break
			}
		}
		retry--
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	}
	if err != nil { //using registry
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
	defer f.Close()
	w.Header().Set("ETag", fmt.Sprintf(`"%s`, ref))
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%.f", client.BlobCacheControlMaxAge.Seconds()))
	if w.Header().Get("Docker-Content-Digest") == "" {
		w.Header().Set("Docker-Content-Digest", ref)
	}
	//MediaType ?
	if w.Header().Get("Content-Type") == "" {
		// Set the content type if not already set.
		w.Header().Set("Content-Type", r.Header["Accept"][0])
	}

	if w.Header().Get("Content-Length") == "" {
		// Set the content length if not already set.
		w.Header().Set("Content-Length", fmt.Sprint(size))
	}
	http.ServeContent(w, r, ref, time.Time{}, f)
}
