package containerd

import (
	"google.golang.org/grpc"
)

type clientOpts struct {
	defaultns   string
	dialOptions []grpc.DialOption
}

// ClientOpt allows callers to set options on the containerd client
type ClientOpt func(c *clientOpts) error