package metadata

import (
	"sync"

	"github.com/llitfkitfk/containerd/snapshots"
)

type snapshotter struct {
	snapshots.Snapshotter
	name string
	db   *DB
	l    sync.RWMutex
}
