package server

import (
	"context"
)

// Server is the containerd main daemon
type Server struct {
}

// New creates and initializes a new containerd server
func New(ctx context.Context, config *Config) (*Server, error) {
	return nil, nil
}

// Stop the containerd server canceling any open connections
func (s *Server) Stop() {
}
