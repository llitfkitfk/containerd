package overlay

import (
	"github.com/llitfkitfk/containerd/platforms"
	"github.com/llitfkitfk/containerd/plugin"
	"github.com/llitfkitfk/containerd/snapshots"
)

func init() {
	plugin.Register(&plugin.Registration{
		Type: plugin.SnapshotPlugin,
		ID:   "overlayfs",
		InitFn: func(ic *plugin.InitContext) (interface{}, error) {
			ic.Meta.Platforms = append(ic.Meta.Platforms, platforms.DefaultSpec())
			ic.Meta.Exports["root"] = ic.Root
			return NewSnapshotter(ic.Root)
		},
	})
}

// NewSnapshotter returns a Snapshotter which uses overlayfs. The overlayfs
// diffs are stored under the provided root. A metadata file is stored under
// the root.
func NewSnapshotter(root string) (snapshots.Snapshotter, error) {
	return nil, nil
}
