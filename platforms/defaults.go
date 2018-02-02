package platforms

import (
	"runtime"

	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

// DefaultSpec returns the current platform's default platform specification.
func DefaultSpec() specs.Platform {
	return specs.Platform{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		// TODO(stevvooe): Need to resolve GOARM for arm hosts.
	}
}
