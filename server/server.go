package server

import (
	"os"
	"path/filepath"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/llitfkitfk/containerd/events/exchange"
	"github.com/llitfkitfk/containerd/log"
	"github.com/llitfkitfk/containerd/plugin"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// New creates and initializes a new containerd server
func New(ctx context.Context, config *Config) (*Server, error) {
	switch {
	case config.Root == "":
		return nil, errors.New("root must be specified")
	case config.State == "":
		return nil, errors.New("state must be specified")
	case config.Root == config.State:
		return nil, errors.New("root and state must be different paths")
	}
	if err := os.MkdirAll(config.Root, 0711); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(config.State, 0711); err != nil {
		return nil, err
	}
	if err := apply(ctx, config); err != nil {
		return nil, err
	}
	plugins, err := loadPlugins(config)
	if err != nil {
		return nil, err
	}
	rpc := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)
	var s = &Server{
		rpc:    rpc,
		events: exchange.NewExchange(),
	}

	for _, p := range plugins {
		id := p.URI()
		log.G(ctx).WithField("type", p.Type).Infof("loading plugin %q...", id)

	}
	return s, nil
}

// Server is the containerd main daemon
type Server struct {
	rpc    *grpc.Server
	events *exchange.Exchange
}

func loadPlugins(config *Config) ([]*plugin.Registration, error) {
	// load all plugins into containerd
	if err := plugin.Load(filepath.Join(config.Root, "plugins")); err != nil {
		return nil, err
	}
	// load additional plugins that don't automatically register themselves
	plugin.Register(&plugin.Registration{
		Type: plugin.ContentPlugin,
		ID:   "content",
		// InitFn: func(ic *plugin.InitContext) (interface{}, error) {
		// 	ic.Meta.Exports["root"] = ic.Root
		// 	return local.NewStore(ic.Root)
		// },
	})
	plugin.Register(&plugin.Registration{
		Type: plugin.MetadataPlugin,
	})
	return nil, nil
}

// Stop the containerd server canceling any open connections
func (s *Server) Stop() {
}

func interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return grpc_prometheus.UnaryServerInterceptor(ctx, req, info, handler)
}
