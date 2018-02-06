package server

import (
	"google.golang.org/grpc/health/grpc_health_v1"
	"expvar"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
	metrics "github.com/docker/go-metrics"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	version "github.com/llitfkitfk/containerd/api/services/version/v1"
	"github.com/llitfkitfk/containerd/content"
	"github.com/llitfkitfk/containerd/content/local"
	"github.com/llitfkitfk/containerd/events/exchange"
	"github.com/llitfkitfk/containerd/log"
	"github.com/llitfkitfk/containerd/metadata"
	"github.com/llitfkitfk/containerd/plugin"
	"github.com/llitfkitfk/containerd/snapshots"
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
	var (
		services []plugin.Service
		s        = &Server{
			rpc:    rpc,
			events: exchange.NewExchange(),
		}
		initialized = plugin.NewPluginSet()
	)
	for _, p := range plugins {
		id := p.URI()
		log.G(ctx).WithField("type", p.Type).Infof("loading plugin %q...", id)

		initContext := plugin.NewContext(
			ctx,
			p,
			initialized,
			config.Root,
			config.State,
		)
		initContext.Events = s.events
		initContext.Address = config.GRPC.Address
		if p.Config != nil {
			pluginConfig, err := config.Decode(p.ID, p.Config)
			if err != nil {
				return nil, err
			}
			initContext.Config = pluginConfig
		}
		result := p.Init(initContext)
		if err := initialized.Add(result); err != nil {
			return nil, errors.Wrapf(err, "could not add plugin result to plugin set")
		}

		instance, err := result.Instance()
		if err != nil {
			if plugin.IsSkipPlugin(err) {
				log.G(ctx).WithField("type", p.Type).Infof("skip loading plugin %q...", id)
			} else {
				log.G(ctx).WithError(err).Warnf("failed to load plugin %s", id)
			}
			continue
		}

		// check for grpc services that should be registered with the server
		if service, ok := instance.(plugin.Service); ok {
			services = append(services, service)
		}
	}

	// register services after all plugins have been initialized
	for _, service := range services {
		if err := service.Register(rpc); err != nil {
			return nil, err
		}
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
		InitFn: func(ic *plugin.InitContext) (interface{}, error) {
			ic.Meta.Exports["root"] = ic.Root
			return local.NewStore(ic.Root)
		},
	})
	plugin.Register(&plugin.Registration{
		Type: plugin.MetadataPlugin,
		ID:   "bolt",
		Requires: []plugin.Type{
			plugin.ContentPlugin,
			plugin.SnapshotPlugin,
		},
		InitFn: func(ic *plugin.InitContext) (interface{}, error) {
			if err := os.MkdirAll(ic.Root, 0711); err != nil {
				return nil, err
			}
			cs, err := ic.Get(plugin.ContentPlugin)
			if err != nil {
				return nil, err
			}

			snapshottersRaw, err := ic.GetByType(plugin.SnapshotPlugin)
			if err != nil {
				return nil, err
			}

			snapshotters := make(map[string]snapshots.Snapshotter)
			for name, sn := range snapshottersRaw {
				sn, err := sn.Instance()
				if err != nil {
					log.G(ic.Context).WithError(err).
						Warnf("could not use snapshotter %v in metadata plugin", name)
					continue
				}
				snapshotters[name] = sn.(snapshots.Snapshotter)
			}

			path := filepath.Join(ic.Root, "meta.db")
			ic.Meta.Exports["path"] = path

			db, err := bolt.Open(path, 0644, nil)
			if err != nil {
				return nil, err
			}
			mdb := metadata.NewDB(db, cs.(content.Store), snapshotters)
			if err := mdb.Init(ic.Context); err != nil {
				return nil, err
			}
			return mdb, nil
		},
	})
	return plugin.Graph(), nil
}

// ServeGRPC provides the containerd grpc APIs on the provided listener
func (s *Server) ServeGRPC(l net.Listener) error {
	// before we start serving the grpc API regster the grpc_prometheus metrics
	// handler.  This needs to be the last service registered so that it can collect
	// metrics for every other service
	grpc_prometheus.Register(s.rpc)
	return trapClosedConnErr(s.rpc.Serve(l))
}

// ServeMetrics provides a prometheus endpoint for exposing metrics
func (s *Server) ServeMetrics(l net.Listener) error {
	m := http.NewServeMux()
	m.Handle("/v1/metrics", metrics.Handler())
	return trapClosedConnErr(http.Serve(l, m))
}

// ServeDebug provides a debug endpoint
func (s *Server) ServeDebug(l net.Listener) error {
	// don't use the default http server mux to make sure nothing gets registered
	// that we don't want to expose via containerd
	m := http.NewServeMux()
	m.Handle("/debug/vars", expvar.Handler())
	m.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	m.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	m.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	m.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	m.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	return trapClosedConnErr(http.Serve(l, m))
}

// Stop the containerd server canceling any open connections
func (s *Server) Stop() {
}

func interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = log.WithModule(ctx, "containerd")
	switch info.Server.(type) {
	case grpc_health_v1.HealthServer:
		// No need to change the context
	case version.VersionServer:
		ctx = log.WithModule(ctx, "version")
	default:
		log.G(ctx).Warnf("unknown GRPC server type: %#v\n", info.Server)
	}
	return grpc_prometheus.UnaryServerInterceptor(ctx, req, info, handler)
}

func trapClosedConnErr(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "use of closed network connection") {
		return nil
	}
	return err
}
