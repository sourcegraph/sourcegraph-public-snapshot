package testsuite

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

var (
	storageKeyName = randomKey()
	storageValue   = fullByteRange()
)

func fullByteRange() (v []byte) {
	for i := byte(0); i < 255; i++ {
		v = append(v, i)
	}
	return
}

func randomKey() string {
	return "my-awesome\x00\x00key" + fmt.Sprint(time.Now().UnixNano())
}

func randomBucket() *sourcegraph.StorageBucket {
	return &sourcegraph.StorageBucket{
		AppName: "go-test",
		Name:    "go-test-bucket" + fmt.Sprint(time.Now().UnixNano()),
		Repo:    "github.com/foo/bar",
	}
}

// Storage_ValidNames tests that valid and complex names are accepted by the
// storage service.
func Storage_ValidNames(ctx context.Context, t *testing.T, s store.Storage) {
	tests := []sourcegraph.StorageKey{
		// Valid bucket name tests.
		sourcegraph.StorageKey{ // Just a normal situation.
			Bucket: &sourcegraph.StorageBucket{
				Name:    "normal-bucket",
				AppName: "normal-app",
			},
			Key: "normal-key",
		},
		sourcegraph.StorageKey{ // Bucket names may contain periods.
			Bucket: &sourcegraph.StorageBucket{
				Name:    "www.sourcegraph.com",
				AppName: "normal-app",
			},
			Key: "normal-key",
		},

		// Valid app name tests.
		sourcegraph.StorageKey{
			Bucket: &sourcegraph.StorageBucket{
				Name:    "my-bucket",
				AppName: "core.app", // App names may contain periods.
				Repo:    "src.sourcegraph.com/foo/bar",
			},
			Key: "normal-key",
		},

		// Valid repo URI tests.
		sourcegraph.StorageKey{ // A normal repo URI.
			Bucket: &sourcegraph.StorageBucket{
				Name:    "www.sourcegraph.com",
				AppName: "normal-app",
				Repo:    "src.sourcegraph.com/foo/bar",
			},
			Key: "normal-key",
		},
		sourcegraph.StorageKey{ // Another normal repo URI.
			Bucket: &sourcegraph.StorageBucket{
				Name:    "www.sourcegraph.com",
				AppName: "normal-app",
				Repo:    "github.com/foo/bar",
			},
			Key: "normal-key",
		},
		sourcegraph.StorageKey{ // Repo "" is accepted as "global storage".
			Bucket: &sourcegraph.StorageBucket{
				Name:    "www.sourcegraph.com",
				AppName: "normal-app",
				Repo:    "",
			},
			Key: "normal-key",
		},
		sourcegraph.StorageKey{ // Crazy null bytes etc. are escaped.
			Bucket: &sourcegraph.StorageBucket{
				Name:    "www.sourcegraph.com",
				AppName: "normal-app",
				Repo:    "src.\x00\x00example.com/\x00/bar",
			},
			Key: "normal-key",
		},
	}

	for _, sk := range tests {
		_, err := s.Put(ctx, &sourcegraph.StoragePutOp{
			Key:   sk,
			Value: storageValue,
		})
		if err != nil {
			t.Fatalf("Put Key: %#q got error=%v\n", sk, err)
		}
	}
}
