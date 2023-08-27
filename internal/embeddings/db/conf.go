package db

import (
	"sync/atomic"

	qdrant "github.com/qdrant/go-client/qdrant"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
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
	}
}

func createCollectionParams(name string, dims uint64, conf *schema.Qdrant) *qdrant.CreateCollection {
	return &qdrant.CreateCollection{
		CollectionName:         name,
		HnswConfig:             getHnswConfigDiff(conf),
		OptimizersConfig:       getOptimizersConfigDiff(conf),
		QuantizationConfig:     getQuantizationConfig(conf),
		OnDiskPayload:          pointers.Ptr(true),
		VectorsConfig:          getVectorsConfig(dims),
		ShardNumber:            nil, // default
		WalConfig:              nil, // default
		ReplicationFactor:      nil, // default
		WriteConsistencyFactor: nil, // default
		InitFromCollection:     nil, // default
	}
}

func updateCollectionParams(name string, conf *schema.Qdrant) *qdrant.UpdateCollection {
	return &qdrant.UpdateCollection{
		CollectionName:     name,
		HnswConfig:         getHnswConfigDiff(conf),
		OptimizersConfig:   getOptimizersConfigDiff(conf),
		QuantizationConfig: getQuantizationConfigDiff(conf),
		Params:             nil, // do not update collection params
		// Do not update vectors config.
		// TODO(camdencheek): consider making OnDisk configurable
		VectorsConfig: nil,
	}
}

func getVectorsConfig(dims uint64) *qdrant.VectorsConfig {
	return &qdrant.VectorsConfig{
		Config: &qdrant.VectorsConfig_Params{
			Params: &qdrant.VectorParams{
				Size:               dims,
				Distance:           qdrant.Distance_Cosine,
				OnDisk:             pointers.Ptr(true),
				HnswConfig:         nil, // use collection default
				QuantizationConfig: nil, // use collection default
			},
		},
	}
}

func getHnswConfigDiff(conf *schema.Qdrant) *qdrant.HnswConfigDiff {
	c := pointers.Deref(conf, schema.Qdrant{})
	overrides := pointers.Deref(c.Hnsw, schema.Hnsw{})

	// Default values should match the documented defaults in site.schema.json.
	return &qdrant.HnswConfigDiff{
		M:                  getUint64(overrides.M, 8),
		PayloadM:           getUint64(overrides.PayloadM, 8),
		EfConstruct:        getUint64(overrides.EfConstruct, 100),
		FullScanThreshold:  getUint64(overrides.FullScanThreshold, 10),
		OnDisk:             pointers.Ptr(pointers.Deref(overrides.OnDisk, true)),
		MaxIndexingThreads: pointers.Ptr(uint64(4)),
	}
}

func getOptimizersConfigDiff(conf *schema.Qdrant) *qdrant.OptimizersConfigDiff {
	c := pointers.Deref(conf, schema.Qdrant{})
	overrides := pointers.Deref(c.Optimizers, schema.Optimizers{})

	// Default values should match the documented defaults in site.schema.json.
	return &qdrant.OptimizersConfigDiff{
		DefaultSegmentNumber:   pointers.Ptr(uint64(4)),
		IndexingThreshold:      getUint64(overrides.IndexingThreshold, 100),
		VacuumMinVectorNumber:  pointers.Ptr(uint64(100)),
		MemmapThreshold:        getUint64(overrides.MemmapThreshold, 100),
		MaxOptimizationThreads: pointers.Ptr(uint64(2)),
	}
}

func getQuantizationConfigDiff(conf *schema.Qdrant) *qdrant.QuantizationConfigDiff {
	c := pointers.Deref(conf, schema.Qdrant{})
	overrides := pointers.Deref(c.Quantization, schema.Quantization{})

	if !pointers.Deref(overrides.Enabled, true) {
		return &qdrant.QuantizationConfigDiff{
			Quantization: &qdrant.QuantizationConfigDiff_Disabled{},
		}
	}

	return &qdrant.QuantizationConfigDiff{
		Quantization: &qdrant.QuantizationConfigDiff_Scalar{
			Scalar: &qdrant.ScalarQuantization{
				Type:      qdrant.QuantizationType_Int8,
				Quantile:  getFloat32(overrides.Quantile, 0.98),
				AlwaysRam: pointers.Ptr(false),
			},
		},
	}
}

func getQuantizationConfig(conf *schema.Qdrant) *qdrant.QuantizationConfig {
	c := pointers.Deref(conf, schema.Qdrant{})
	overrides := pointers.Deref(c.Quantization, schema.Quantization{})

	if !pointers.Deref(overrides.Enabled, true) {
		return nil
	}
	return &qdrant.QuantizationConfig{
		Quantization: &qdrant.QuantizationConfig_Scalar{
			Scalar: &qdrant.ScalarQuantization{
				Type:      qdrant.QuantizationType_Int8,
				Quantile:  getFloat32(overrides.Quantile, 0.95),
				AlwaysRam: pointers.Ptr(false),
			},
		},
	}
}

func getUint64(input *int, def uint64) *uint64 {
	if input != nil {
		u := uint64(*input)
		return &u
	}
	return &def
}

func getFloat32(input *float64, def float32) *float32 {
	if input != nil {
		f := float32(*input)
		return &f
	}
	return &def
}
