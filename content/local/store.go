package local

import (
	"os"
	"path/filepath"

	"github.com/llitfkitfk/containerd/content"
	digest "github.com/opencontainers/go-digest"
)

// LabelStore is used to store mutable labels for digests
type LabelStore interface {
	// Get returns all the labels for the given digest
	Get(digest.Digest) (map[string]string, error)

	// Set sets all the labels for a given digest
	Set(digest.Digest, map[string]string) error

	// Update replaces the given labels for a digest,
	// a key with an empty value removes a label.
	Update(digest.Digest, map[string]string) (map[string]string, error)
}

// Store is digest-keyed store for content. All data written into the store is
// stored under a verifiable digest.
//
// Store can generally support multi-reader, single-writer ingest of data,
// including resumable ingest.
type store struct {
	root string
	ls   LabelStore
}

// NewStore returns a local content store
func NewStore(root string) (content.Store, error) {
	return NewLabeledStore(root, nil)
}

// NewLabeledStore returns a new content store using the provided label store
//
// Note: content stores which are used underneath a metadata store may not
// require labels and should use `NewStore`. `NewLabeledStore` is primarily
// useful for tests or standalone implementations.
func NewLabeledStore(root string, ls LabelStore) (content.Store, error) {
	if err := os.MkdirAll(filepath.Join(root, "ingest"), 0777); err != nil {
		return nil, err
	}

	return &store{
		root: root,
		ls:   ls,
	}, nil
}
