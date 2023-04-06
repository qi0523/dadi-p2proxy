package client

import (
	"context"
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
	RETRY                  = 3
	BlobCacheControlMaxAge = 365 * 24 * time.Hour
)

var client *containerd.Client

func init() {
	var err error
	client, err = containerd.New(ContainerdSockPath, containerd.WithDefaultNamespace(ContainerdNameSpace))
	if err != nil {
		logrus.Fatal("failed to create containerd client.")
	}
}

func GetManifestInfoByTag(name, tag string) (string, digest.Digest, error) {
	images, err := client.ImageService().List(context.Background())
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
