package main

import (
	"github.com/llitfkitfk/containerd/defaults"
	"github.com/llitfkitfk/containerd/server"
)

func defaultConfig() *server.Config {
	return &server.Config{
		Root:  defaults.DefaultRootDir,
		State: defaults.DefaultStateDir,
		GRPC: server.GRPCConfig{
			Address: defaults.DefaultAddress,
		},
		NoSubreaper: false,
		Debug: server.Debug{
			Level:   "info",
			Address: defaults.DefaultDebugAddress,
		},
	}
}
