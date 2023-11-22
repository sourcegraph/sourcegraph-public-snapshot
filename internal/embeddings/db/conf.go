package db

import (
	"sync/atomic"

	qdrant "github.com/qdrant/go-client/qdrant"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"google.golang.org/grpc"
)

// NewDBFromConfFunc returns a function that can be called to get an
// VectorDB instance based on the connection info from the conf package.
// It will watch conf and update the connection if there are any changes.
//
// If Qdrant is disabled, it will instead return the provided default VectorDB.
func NewDBFromConfFunc(logger log.Logger, def VectorDB) func() (VectorDB, error) {
	type connAndErr struct {
		conn *grpc.ClientConn
		err  error
	}
	var (
		oldAddr string
		ptr     atomic.Pointer[connAndErr]
	)

	conf.Watch(func() {
		c := conf.Get()
		qc := conf.GetEmbeddingsConfig(c.SiteConfiguration)
		if qc == nil || !qc.Qdrant.Enabled {
			// Embeddings is disabled. Clear any errors and close any previous connection.
			old := ptr.Swap(&connAndErr{nil, nil})
			if old != nil && old.conn != nil {
				old.conn.Close()
			}
			return
		}
		if newAddr := c.ServiceConnections().Qdrant; newAddr != oldAddr {
			// The address has changed or this is running for the first time.
			// Attempt to open dial Qdrant.
			newConn, newErr := defaults.Dial(newAddr, logger)
			oldAddr = newAddr
			old := ptr.Swap(&connAndErr{newConn, newErr})
			if old != nil && old.conn != nil {
				old.conn.Close()
			}
		}
	})

	return func() (VectorDB, error) {
		curr := ptr.Load()
		if curr == nil {
			return def, nil
		}
		if curr.err != nil {
			return nil, curr.err
		}
		if curr.conn == nil {
			return def, nil
		}

		return NewQdrantDBFromConn(curr.conn), nil
	}
}

func createCollectionParams(name string, dims uint64, qc conftypes.QdrantConfig) *qdrant.CreateCollection {
	return &qdrant.CreateCollection{
		CollectionName:         name,
		HnswConfig:             getHnswConfigDiff(qc.QdrantHNSWConfig),
		OptimizersConfig:       getOptimizersConfigDiff(qc.QdrantOptimizersConfig),
		QuantizationConfig:     getQuantizationConfig(qc.QdrantQuantizationConfig),
		OnDiskPayload:          pointers.Ptr(true),
		VectorsConfig:          getVectorsConfig(dims),
		ShardNumber:            nil, // default
		WalConfig:              nil, // default
		ReplicationFactor:      nil, // default
		WriteConsistencyFactor: nil, // default
		InitFromCollection:     nil, // default
	}
}

func updateCollectionParams(name string, qc conftypes.QdrantConfig) *qdrant.UpdateCollection {
	return &qdrant.UpdateCollection{
		CollectionName:     name,
		HnswConfig:         getHnswConfigDiff(qc.QdrantHNSWConfig),
		OptimizersConfig:   getOptimizersConfigDiff(qc.QdrantOptimizersConfig),
		QuantizationConfig: getQuantizationConfigDiff(qc.QdrantQuantizationConfig),
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

func getHnswConfigDiff(qhc conftypes.QdrantHNSWConfig) *qdrant.HnswConfigDiff {
	return &qdrant.HnswConfigDiff{
		M:                  qhc.M,
		PayloadM:           qhc.PayloadM,
		EfConstruct:        qhc.EfConstruct,
		FullScanThreshold:  qhc.FullScanThreshold,
		OnDisk:             &qhc.OnDisk,
		MaxIndexingThreads: nil, // default
	}
}

func getOptimizersConfigDiff(qoc conftypes.QdrantOptimizersConfig) *qdrant.OptimizersConfigDiff {
	// Default values should match the documented defaults in site.schema.json.
	return &qdrant.OptimizersConfigDiff{
		IndexingThreshold:      &qoc.IndexingThreshold,
		MemmapThreshold:        &qoc.MemmapThreshold,
		DefaultSegmentNumber:   nil,
		VacuumMinVectorNumber:  nil,
		MaxOptimizationThreads: nil,
	}
}

func getQuantizationConfigDiff(qqc conftypes.QdrantQuantizationConfig) *qdrant.QuantizationConfigDiff {
	if !qqc.Enabled {
		return &qdrant.QuantizationConfigDiff{
			Quantization: &qdrant.QuantizationConfigDiff_Disabled{},
		}
	}

	return &qdrant.QuantizationConfigDiff{
		Quantization: &qdrant.QuantizationConfigDiff_Scalar{
			Scalar: &qdrant.ScalarQuantization{
				Type:      qdrant.QuantizationType_Int8,
				Quantile:  &qqc.Quantile,
				AlwaysRam: pointers.Ptr(false),
			},
		},
	}
}

func getQuantizationConfig(qqc conftypes.QdrantQuantizationConfig) *qdrant.QuantizationConfig {
	if !qqc.Enabled {
		return nil
	}
	return &qdrant.QuantizationConfig{
		Quantization: &qdrant.QuantizationConfig_Scalar{
			Scalar: &qdrant.ScalarQuantization{
				Type:      qdrant.QuantizationType_Int8,
				Quantile:  &qqc.Quantile,
				AlwaysRam: pointers.Ptr(false),
			},
		},
	}
}
