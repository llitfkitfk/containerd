package containerd

import (
	"context"
	"fmt"
	"runtime"
	"time"

	ptypes "github.com/gogo/protobuf/types"
	versionservice "github.com/llitfkitfk/containerd/api/services/version/v1"
	"github.com/llitfkitfk/containerd/dialer"
	"github.com/llitfkitfk/containerd/plugin"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// New returns a new containerd client that is connected to the containerd
// instance provided by address
func New(address string, opts ...ClientOpt) (*Client, error) {
	var copts clientOpts
	for _, o := range opts {
		if err := o(&copts); err != nil {
			return nil, err
		}
	}
	gopts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithTimeout(10 * time.Second),
		grpc.FailOnNonTempDialError(true),
		grpc.WithBackoffMaxDelay(3 * time.Second),
		grpc.WithDialer(dialer.Dialer),
	}
	if len(copts.dialOptions) > 0 {
		gopts = copts.dialOptions
	}

	conn, err := grpc.Dial(dialer.DialAddress(address), gopts...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to dial %q", address)
	}
	return NewWithConn(conn, opts...)
}

// NewWithConn returns a new containerd client that is connected to the containerd
// instance provided by the connection
func NewWithConn(conn *grpc.ClientConn, opts ...ClientOpt) (*Client, error) {
	return &Client{
		conn:    conn,
		runtime: fmt.Sprintf("%s.%s", plugin.RuntimePlugin, runtime.GOOS),
	}, nil
}

// Client is the client to interact with containerd and its various services
// using a uniform interface
type Client struct {
	conn    *grpc.ClientConn
	runtime string
}

// VersionService returns the underlying VersionClient
func (c *Client) VersionService() versionservice.VersionClient {
	return versionservice.NewVersionClient(c.conn)
}

// Version of containerd
type Version struct {
	// Version number
	Version string
	// Revision from git that was built
	Revision string
}

// Version returns the version of containerd that the client is connected to
func (c *Client) Version(ctx context.Context) (Version, error) {
	response, err := c.VersionService().Version(ctx, &ptypes.Empty{})
	if err != nil {
		return Version{}, err
	}
	return Version{
		Version:  response.Version,
		Revision: response.Revision,
	}, nil
}

// Close closes the clients connection to containerd
func (c *Client) Close() error {
	return c.conn.Close()
}
