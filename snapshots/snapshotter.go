package snapshots

import (
	"context"
	"time"
)

// Kind identifies the kind of snapshot.
type Kind uint8

// Info provides information about a particular snapshot.
// JSON marshallability is supported for interactive with tools like ctr,
type Info struct {
	Kind    Kind              // active or committed snapshot
	Name    string            // name or key of snapshot
	Parent  string            `json:",omitempty"` // name of parent snapshot
	Labels  map[string]string `json:",omitempty"` // Labels for snapshot
	Created time.Time         `json:",omitempty"` // Created time
	Updated time.Time         `json:",omitempty"` // Last update time
}

// Snapshotter defines the methods required to implement a snapshot snapshotter for
// allocating, snapshotting and mounting filesystem changesets. The model works
// by building up sets of changes with parent-child relationships.
//
// A snapshot represents a filesystem state. Every snapshot has a parent, where
// the empty parent is represented by the empty string. A diff can be taken
// between a parent and its snapshot to generate a classic layer.
//
// An active snapshot is created by calling `Prepare`. After mounting, changes
// can be made to the snapshot. The act of committing creates a committed
// snapshot. The committed snapshot will get the parent of active snapshot. The
// committed snapshot can then be used as a parent. Active snapshots can never
// act as a parent.
//
// Snapshots are best understood by their lifecycle. Active snapshots are
// always created with Prepare or View. Committed snapshots are always created
// with Commit.  Active snapshots never become committed snapshots and vice
// versa. All snapshots may be removed.
//
// For consistency, we define the following terms to be used throughout this
// interface for snapshotter implementations:
//
// 	`ctx` - refers to a context.Context
// 	`key` - refers to an active snapshot
// 	`name` - refers to a committed snapshot
// 	`parent` - refers to the parent in relation
//
// Most methods take various combinations of these identifiers. Typically,
// `name` and `parent` will be used in cases where a method *only* takes
// committed snapshots. `key` will be used to refer to active snapshots in most
// cases, except where noted. All variables used to access snapshots use the
// same key space. For example, an active snapshot may not share the same key
// with a committed snapshot.
//
// We cover several examples below to demonstrate the utility of a snapshot
// snapshotter.
//
// Importing a Layer
//
// To import a layer, we simply have the Snapshotter provide a list of
// mounts to be applied such that our dst will capture a changeset. We start
// out by getting a path to the layer tar file and creating a temp location to
// unpack it to:
//
//	layerPath, tmpDir := getLayerPath(), mkTmpDir() // just a path to layer tar file.
//
// We start by using a Snapshotter to Prepare a new snapshot transaction, using a
// key and descending from the empty parent "":
//
//	mounts, err := snapshotter.Prepare(ctx, key, "")
// 	if err != nil { ... }
//
// We get back a list of mounts from Snapshotter.Prepare, with the key identifying
// the active snapshot. Mount this to the temporary location with the
// following:
//
//	if err := mount.All(mounts, tmpDir); err != nil { ... }
//
// Once the mounts are performed, our temporary location is ready to capture
// a diff. In practice, this works similar to a filesystem transaction. The
// next step is to unpack the layer. We have a special function unpackLayer
// that applies the contents of the layer to target location and calculates the
// DiffID of the unpacked layer (this is a requirement for docker
// implementation):
//
//	layer, err := os.Open(layerPath)
//	if err != nil { ... }
// 	digest, err := unpackLayer(tmpLocation, layer) // unpack into layer location
// 	if err != nil { ... }
//
// When the above completes, we should have a filesystem the represents the
// contents of the layer. Careful implementations should verify that digest
// matches the expected DiffID. When completed, we unmount the mounts:
//
//	unmount(mounts) // optional, for now
//
// Now that we've verified and unpacked our layer, we commit the active
// snapshot to a name. For this example, we are just going to use the layer
// digest, but in practice, this will probably be the ChainID:
//
//	if err := snapshotter.Commit(ctx, digest.String(), key); err != nil { ... }
//
// Now, we have a layer in the Snapshotter that can be accessed with the digest
// provided during commit. Once you have committed the snapshot, the active
// snapshot can be removed with the following:
//
// 	snapshotter.Remove(ctx, key)
//
// Importing the Next Layer
//
// Making a layer depend on the above is identical to the process described
// above except that the parent is provided as parent when calling
// Manager.Prepare, assuming a clean, unique key identifier:
//
// 	mounts, err := snapshotter.Prepare(ctx, key, parentDigest)
//
// We then mount, apply and commit, as we did above. The new snapshot will be
// based on the content of the previous one.
//
// Running a Container
//
// To run a container, we simply provide Snapshotter.Prepare the committed image
// snapshot as the parent. After mounting, the prepared path can
// be used directly as the container's filesystem:
//
// 	mounts, err := snapshotter.Prepare(ctx, containerKey, imageRootFSChainID)
//
// The returned mounts can then be passed directly to the container runtime. If
// one would like to create a new image from the filesystem, Manager.Commit is
// called:
//
// 	if err := snapshotter.Commit(ctx, newImageSnapshot, containerKey); err != nil { ... }
//
// Alternatively, for most container runs, Snapshotter.Remove will be called to
// signal the Snapshotter to abandon the changes.
type Snapshotter interface {
	// Stat returns the info for an active or committed snapshot by name or
	// key.
	//
	// Should be used for parent resolution, existence checks and to discern
	// the kind of snapshot.
	Stat(ctx context.Context, key string) (Info, error)
}
