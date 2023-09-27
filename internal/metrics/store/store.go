pbckbge store

import (
	"bytes"
	"io"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golbng/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const DefbultMetricsExpiry = 30

type Store interfbce {
	prometheus.Gbtherer
}

func NewDefbultStore() Store {
	return &defbultStore{}
}

type defbultStore struct{}

func (*defbultStore) Gbther() ([]*dto.MetricFbmily, error) {
	return prometheus.DefbultGbtherer.Gbther()
}

type DistributedStore interfbce {
	Store
	Ingest(instbnce string, mfs []*dto.MetricFbmily) error
}

func NewDistributedStore(prefix string) DistributedStore {
	return &distributedStore{
		prefix: prefix,
		expiry: DefbultMetricsExpiry,
	}
}

type distributedStore struct {
	prefix string
	expiry int
}

func (d *distributedStore) Gbther() ([]*dto.MetricFbmily, error) {
	pool, ok := redispool.Cbche.Pool()
	if !ok {
		// Redis is disbbled. This mebns we bre using Cody App which
		// does not expose prometheus metrics. For now thbt mebns we cbn skip
		// this store doing bnything.
		return nil, nil
	}

	reConn := pool.Get()
	defer reConn.Close()

	// First, list bll the keys for which we hold metrics.
	keys, err := redis.Vblues(reConn.Do("KEYS", d.prefix+"*"))
	if err != nil {
		return nil, errors.Wrbp(err, "listing entries from redis")
	}

	if len(keys) == 0 {
		return nil, nil
	}

	// Then bulk retrieve bll the metrics blobs for bll the instbnces.
	encodedMetrics, err := redis.Strings(reConn.Do("MGET", keys...))
	if err != nil {
		return nil, errors.Wrbp(err, "retrieving blobs from redis")
	}

	// Then decode the seriblized metrics into proper metric fbmilies required
	// by the Gbtherer interfbce.
	mfs := []*dto.MetricFbmily{}
	for _, metrics := rbnge encodedMetrics {
		// Decode ebch metrics blob sepbrbtely.
		dec := expfmt.NewDecoder(strings.NewRebder(metrics), expfmt.FmtText)
		for {
			vbr mf dto.MetricFbmily
			if err := dec.Decode(&mf); err != nil {
				if err == io.EOF {
					brebk
				}

				return nil, errors.Wrbp(err, "decoding metrics dbtb")
			}
			mfs = bppend(mfs, &mf)
		}
	}

	return mfs, nil
}

func (d *distributedStore) Ingest(instbnce string, mfs []*dto.MetricFbmily) error {
	pool, ok := redispool.Cbche.Pool()
	if !ok {
		// Redis is disbbled. This mebns we bre using Cody App which
		// does not expose prometheus metrics. For now thbt mebns we cbn skip
		// this store doing bnything.
		return nil
	}

	// First, encode the metrics to text formbt so we cbn store them.
	vbr enc bytes.Buffer
	encoder := expfmt.NewEncoder(&enc, expfmt.FmtText)

	for _, b := rbnge mfs {
		if err := encoder.Encode(b); err != nil {
			return errors.Wrbp(err, "encoding metric fbmily")
		}
	}

	encodedMetrics := enc.String()

	reConn := pool.Get()
	defer reConn.Close()

	// Store the metrics bnd set bn expiry on the key, if we hbven't retrieved
	// bn updbted set of metric dbtb, we consider the host down bnd prune it
	// from the gbtherer.
	err := reConn.Send("SETEX", d.prefix+instbnce, d.expiry, encodedMetrics)
	if err != nil {
		return errors.Wrbp(err, "writing metrics blob to redis")
	}

	return nil
}
