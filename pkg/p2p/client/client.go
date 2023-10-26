package client

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/containerd/containerd"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
)

const (
	ContainerdRoot         = "/var/lib/containerd/io.containerd.content.v1.content"
	ContainerdSockPath     = "/run/containerd/containerd.sock"
	ContainerdNameSpace    = "default"
	RETRY                  = 10
	RAND_INTN              = 15
	BlobCacheControlMaxAge = 365 * 24 * time.Hour
)

var client *containerd.Client

var hostIp string

func GetContainerdClient() *containerd.Client {
	if client == nil {
		var err error
		client, err = containerd.New(ContainerdSockPath, containerd.WithDefaultNamespace(ContainerdNameSpace))
		if err != nil {
			logrus.Fatal("failed to create containerd client.")
		}
	}
	return client
}

func GetManifestInfoByTag(c *containerd.Client, name, tag string) (string, digest.Digest, error) {
	images, err := c.ImageService().List(context.Background())
	if err != nil {
		return "", "", err
	}
	for _, image := range images {
		if image.Name[strings.Index(image.Name, "/")+1:] == name+":"+tag {
			return image.Target.MediaType, image.Target.Digest, nil
		}
	}
	return "", "", nil
}

func GetManifestInfoByTmpImage(c *containerd.Client, name, tag string) (string, digest.Digest, int64, error) {
	tmpimage, err := c.GetTmpImage(context.Background(), name+":"+tag)
	if err != nil {
		return "", "", 0, err
	}
	return tmpimage.Target.MediaType, tmpimage.Target.Digest, tmpimage.Target.Size, nil
}

func GetOutBoundIP() (ip string, err error) {
	if hostIp != "" {
		return hostIp, nil
	}
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return "", err
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	// fmt.Println(localAddr.String())
	hostIp = strings.Split(localAddr.String(), ":")[0]
	return hostIp, nil
}
