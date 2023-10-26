package metagc

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/data-accelerator/dadi-p2proxy/pkg/p2p/client"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var grpcClient MetaGcServiceClient

var gcTask *GcTask

type GcTask struct {
	data             map[string]map[string]struct{}
	mu               sync.Mutex
	containerdClient *containerd.Client
}

func (t *GcTask) Insert(funcName string, imageName string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	set, ok := t.data[imageName]

	if !ok {
		set = make(map[string]struct{})
		set[funcName] = struct{}{}
		t.data[imageName] = set
		return
	}

	if _, ok = set[funcName]; !ok {
		set[funcName] = struct{}{}
		return
	}
}

func (t *GcTask) getGcInfo(imageName string) string {
	if set, ok := t.data[imageName]; ok {
		if len(set) == 0 {
			return ""
		}
		var res string

		for key, _ := range set {
			res += key
			res += "|"
		}
		return res
	}
	return ""
}

func (t *GcTask) Run() {
	ctx := context.Background()
	for {

		images, err := t.containerdClient.ImageService().List(ctx)

		if err != nil {
			panic(err)
		}

		iNames := make(map[string]struct{}, 0)

		for _, image := range images {
			iNames[image.Name] = struct{}{}
		}

		cons, err := t.containerdClient.ContainerService().List(ctx)

		if err != nil {
			panic(err)
		}

		for _, con := range cons {
			delete(iNames, con.Image)
		}

		t.mu.Lock()

		if len(iNames) != 0 {
			var res string
			for key, _ := range iNames {
				infos := t.getGcInfo(key)
				if infos == "" {
					continue
				}
				res += infos
			}
			if res != "" {
				res = res[:len(res)-1]
				hostIp, err := client.GetOutBoundIP()
				if err != nil {
					panic(err)
				}
				req := MetaGcRequest{ActionName: res, Kind: res, InvokerIp: hostIp}
				retries := 3
				for retries > 0 {
					if _, err := GetGRPCClient().GcMetadata(ctx, &req); err == nil {
						for iName, _ := range iNames {
							t.containerdClient.ImageService().Delete(ctx, iName)
							delete(t.data, iName)
						}
						break
					}
					retries--
				}
			}
		}
		t.mu.Unlock()
		time.Sleep(5 * time.Minute)
	}
}

func GetGRPCClient() MetaGcServiceClient {
	if grpcClient == nil {
		conn, err := grpc.Dial(os.Args[3]+":13002", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic("grpc dial failed.")
		}
		grpcClient = NewMetaGcServiceClient(conn)
	}
	return grpcClient
}

func GetGcTask() *GcTask {
	if gcTask == nil {
		gcTask = &GcTask{
			data:             make(map[string]map[string]struct{}),
			containerdClient: client.GetContainerdClient(),
		}

		go gcTask.Run()
		return gcTask
	}
	return gcTask
}
