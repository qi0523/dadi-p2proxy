package metagc

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/data-accelerator/dadi-p2proxy/pkg/p2p/client"
	"github.com/sirupsen/logrus"
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
	if funcName == "" {
		return
	}
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

func (t *GcTask) gc_handler(iNames map[string]struct{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	logrus.Info("gc_handler is start...")
	if len(iNames) != 0 {
		var res string
		for key, _ := range iNames {
			infos := t.getGcInfo(key)
			if infos == "" {
				continue
			}
			res += infos
		}
		logrus.Info("gc res, ", res)
		if res != "" {
			res = res[:len(res)-1]
			hostIp, err := client.GetOutBoundIP()
			if err != nil {
				logrus.Info("hostIp can not be got.")
				return
			}
			req := MetaGcRequest{ActionName: res, Kind: res, InvokerIp: hostIp}
			retries := 3
			ctx := context.Background()
			for retries > 0 {
				if _, err := GetGRPCClient().GcMetadata(ctx, &req); err == nil {
					for iName, _ := range iNames {
						t.containerdClient.ImageService().Delete(ctx, iName)
						delete(t.data, iName)
						logrus.Info("gc_hander, image is deleted, ", iName)
					}
					break
				}
				retries--
			}
		}
	}
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
			logrus.Info("imageList ", image.Name)
		}

		cons, err := t.containerdClient.ContainerService().List(ctx)

		if err != nil {
			panic(err)
		}

		for _, con := range cons {
			delete(iNames, con.Image)
			logrus.Info("containerList ", con.Image)
		}

		t.gc_handler(iNames)

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
		logrus.Info("gc is running....")
		return gcTask
	}
	return gcTask
}
