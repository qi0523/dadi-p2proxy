package fs

import "os"

var token string

func getToken() string {
	if token == "" {
		// get from /etc/containerd/certs.d/hosts.toml
		token = os.Args[4]
		if token == "" {
			panic("token is empty")
		}
		return token
	}
	return token
}