package metadata

import (
	"context"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/llitfkitfk/containerd/content"
	"github.com/llitfkitfk/containerd/snapshots"
)

// DB represents a metadata database backed by a bolt
// database. The database is fully namespaced and stores
// image, container, namespace, snapshot, and content data
// while proxying data shared across namespaces to backend
// datastores for content and snapshots.
type DB struct {
	db *bolt.DB
	ss map[string]*snapshotter

	// wlock is used to protect access to the data structures during garbage
	// collection. While the wlock is held no writable transactions can be
	// opened, preventing changes from occurring between the mark and
	// sweep phases without preventing read transactions.
	wlock sync.RWMutex

	// dirty flags and lock keeps track of datastores which have had deletions
	// since the last garbage collection. These datastores will will be garbage
	// collected during the next garbage collection.
	dirtyL  sync.Mutex
	dirtySS map[string]struct{}
	dirtyCS bool

	// mutationCallbacks are called after each mutation with the flag
	// set indicating whether any dirty flags are set
	mutationCallbacks []func(bool)
}

// NewDB creates a new metadata database using the provided
// bolt database, content store, and snapshotters.
func NewDB(db *bolt.DB, cs content.Store, ss map[string]snapshots.Snapshotter) *DB {
	m := &DB{
		db:      db,
		ss:      make(map[string]*snapshotter, len(ss)),
		dirtySS: map[string]struct{}{},
	}

	return m
}

// Init ensures the database is at the correct version
// and performs any needed migrations.
func (m *DB) Init(ctx context.Context) error {
	return nil
}
