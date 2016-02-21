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

// Storage_InvalidNames tests that invalid names are not allowed by the storage
// service.
func Storage_InvalidNames(ctx context.Context, t *testing.T, s store.Storage) {
	tests := []sourcegraph.StorageBucket{
		// Invalid bucket name tests.
		sourcegraph.StorageBucket{
			Name:    " startswithspace",
			AppName: "my-app",
			Repo:    "src.sourcegraph.com/foo/bar",
		},
		sourcegraph.StorageBucket{
			Name:    "endswithspace ",
			AppName: "my-app",
			Repo:    "src.sourcegraph.com/foo/bar",
		},
		sourcegraph.StorageBucket{
			Name:    "contains space",
			AppName: "my-app",
			Repo:    "src.sourcegraph.com/foo/bar",
		},

		// Invalid app name tests.
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: " startswithspace",
			Repo:    "src.sourcegraph.com/foo/bar",
		},
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "endswithspace ",
			Repo:    "src.sourcegraph.com/foo/bar",
		},
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "contains space",
			Repo:    "src.sourcegraph.com/foo/bar",
		},

		// Invalid repo URI tests.
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "my-app",
			Repo:    " starts.with.space/foo/bar",
		},
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "my-app",
			Repo:    "ends.with.space/foo/bar ",
		},
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "my-app",
			Repo:    "http://src.sourcegraph.com/foo/bar", // scheme not allowed
		},
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "my-app",
			Repo:    "src.sourcegraph.com/foo/bar?ok=true", // query not allowed
		},
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "my-app",
			Repo:    "src.sourcegraph.com/foo/bar#ok", // fragment not allowed
		},
	}

	for _, bucket := range tests {
		_, err := s.Put(ctx, &sourcegraph.StoragePutOp{
			Key: sourcegraph.StorageKey{
				Bucket: &bucket,
				Key:    storageKeyName,
			},
			Value: storageValue,
		})
		if err == nil {
			t.Logf("Put Key.Bucket: %#q\n", bucket)
			t.Fatal("expected error for non-compliant bucket name")
		}
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
