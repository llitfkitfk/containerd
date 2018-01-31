package main

import (
	"context"
	"os"

	"github.com/llitfkitfk/containerd/log"
	"github.com/llitfkitfk/containerd/server"
	"golang.org/x/sys/unix"
)

const defaultConfigPath = "/etc/containerd/config.toml"

var handledSignals = []os.Signal{
	unix.SIGTERM,
	unix.SIGINT,
}

func handleSignals(ctx context.Context, signals chan os.Signal, serverC chan *server.Server) chan struct{} {
	done := make(chan struct{}, 1)
	go func() {
		var server *server.Server
		for {
			select {
			case s := <-serverC:
				server = s
			case s := <-signals:
				log.G(ctx).WithField("signal", s).Debug("received signal")
				switch s {
				default:
					if server == nil {
						close(done)
						return
					}
					server.Stop()
					close(done)
				}
			}
		}
	}()
	return done
}
