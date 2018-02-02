package naive

import (
	"context"

	"github.com/llitfkitfk/containerd/platforms"
	"github.com/llitfkitfk/containerd/plugin"
	"github.com/llitfkitfk/containerd/snapshots"
)

func init() {
	plugin.Register(&plugin.Registration{
		Type: plugin.SnapshotPlugin,
		ID:   "naive",
		InitFn: func(ic *plugin.InitContext) (interface{}, error) {
			ic.Meta.Platforms = append(ic.Meta.Platforms, platforms.DefaultSpec())
			return NewSnapshotter(ic.Root)
		},
	})
}

type snapshotter struct {
	root string
}

// NewSnapshotter returns a Snapshotter which copies layers on the underlying
// file system. A metadata file is stored under the root.
func NewSnapshotter(root string) (snapshots.Snapshotter, error) {
	return &snapshotter{
		root: root,
	}, nil
}

// Stat returns the info for an active or committed snapshot by name or
// key.
//
// Should be used for parent resolution, existence checks and to discern
// the kind of snapshot.
func (o *snapshotter) Stat(ctx context.Context, key string) (snapshots.Info, error) {
	return snapshots.Info{}, nil
}
