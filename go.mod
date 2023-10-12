module github.com/data-accelerator/dadi-p2proxy

go 1.15

require (
	github.com/containerd/containerd v1.7.6
	github.com/dgraph-io/ristretto v0.1.1
	github.com/elazarl/goproxy v0.0.0-20230808193330-2592e75ae04a
	github.com/opencontainers/go-digest v1.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.17.0
	github.com/stretchr/testify v1.8.4
	google.golang.org/grpc v1.58.2
	google.golang.org/protobuf v1.31.0
)

replace (
	github.com/containerd/containerd => github.com/qi0523/containerd v1.6.9
	github.com/elazarl/goproxy => github.com/taoting1234/goproxy v0.0.0-20210901033843-ebf581737889
)
