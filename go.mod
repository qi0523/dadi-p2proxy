module github.com/data-accelerator/dadi-p2proxy

go 1.15

require (
	github.com/containerd/containerd v1.7.0
	github.com/dgraph-io/ristretto v0.1.0
	github.com/elazarl/goproxy v0.0.0-20210801061803-8e322dfb79c4
	github.com/opencontainers/go-digest v1.0.0
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.6.0
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.8.2
)

replace github.com/elazarl/goproxy => github.com/taoting1234/goproxy v0.0.0-20210901033843-ebf581737889
