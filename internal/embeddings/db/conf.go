package db

import (
	"sync/atomic"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"google.golang.org/grpc"
)

// NewDBFromConfFunc returns a function that can be called to get an
// VectorDB instance based on the connection info from the conf package.
// It will watch conf and update the connection if there are any changes.
//
// If Qdrant is disabled, it will instead return the provided default VectorDB.
func NewDBFromConfFunc(logger log.Logger, def VectorDB) func() (VectorDB, error) {
	var (
		oldAddr string
		err     error
		ptr     atomic.Pointer[grpc.ClientConn]
	)

	conf.Watch(func() {
		if newAddr := conf.Get().ServiceConnections().Qdrant; newAddr != oldAddr {
			newConn, dialErr := defaults.Dial(newAddr, logger)
			oldAddr = newAddr
			err = dialErr
			oldConn := ptr.Swap(newConn)
			if oldConn != nil {
				oldConn.Close()
			}
		}
	})

	return func() (VectorDB, error) {
		if err != nil {
			return nil, err
		}

		conn := ptr.Load()
		if conn == nil {
			return def, nil
		}

		return NewQdrantDBFromConn(conn), nil
=======
	qdrant "github.com/qdrant/go-client/qdrant"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func getHnswConfigDiff(conf *schema.QdrantConfig) *qdrant.HnswConfigDiff {
	if conf == nil {
		return nil
	}
	return &qdrant.HnswConfigDiff{
		M:                 pointers.Ptr(uint64(conf.Hnsw.M)),
		EfConstruct:       pointers.Ptr(uint64(conf.Hnsw.EfConstruct)),
		FullScanThreshold: pointers.Ptr(uint64(conf.Hnsw.FullScanThreshold)),
		OnDisk:            pointers.Ptr(conf.Hnsw.OnDisk),
		PayloadM:          pointers.Ptr(uint64(conf.Hnsw.PayloadM)),
	}
}

func getOptimizersConfigDiff(conf *schema.QdrantConfig) *qdrant.OptimizersConfigDiff {
	if conf == nil {
		return nil
	}
	return &qdrant.OptimizersConfigDiff{
		IndexingThreshold: pointers.Ptr(uint64(conf.Optimizers.IndexingThreshold)),
		MemmapThreshold:   pointers.Ptr(uint64(conf.Optimizers.MemmapThreshold)),
	}
}

func getQuantizationConfigDiff(conf *schema.QdrantConfig) *qdrant.QuantizationConfigDiff {
	if conf == nil {
		return nil
	}
	if !conf.Quantization.Enabled {
		return &qdrant.QuantizationConfigDiff{
			Quantization: &qdrant.QuantizationConfigDiff_Disabled{},
		}
	}
	return &qdrant.QuantizationConfigDiff{
		Quantization: &qdrant.QuantizationConfigDiff_Scalar{
			Scalar: &qdrant.ScalarQuantization{
				Type:      qdrant.QuantizationType_Int8,
				Quantile:  pointers.Ptr(float32(conf.Quantization.Quantile)),
				AlwaysRam: pointers.Ptr(false),
			},
		},
	}
}

func getQuantizationConfig(conf *schema.QdrantConfig) *qdrant.QuantizationConfig {
	if !conf.Quantization.Enabled {
		return nil
	}
	return &qdrant.QuantizationConfig{
		Quantization: &qdrant.QuantizationConfig_Scalar{
			Scalar: &qdrant.ScalarQuantization{
				Type:      qdrant.QuantizationType_Int8,
				Quantile:  pointers.Ptr(float32(conf.Quantization.Quantile)),
				AlwaysRam: pointers.Ptr(false),
			},
		},
	}
}
