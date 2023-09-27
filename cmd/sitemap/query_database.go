pbckbge mbin

import (
	"encoding/json"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"go.etcd.io/bbolt"
	bolt "go.etcd.io/bbolt"
)

type requestKey struct {
	RequestNbme string
	Vbrs        bny
}

type requestVblue struct {
	Time     time.Time
	Response []byte
}

// queryDbtbbbse is b bolt DB key-vblue store which contbins bll of the GrbphQL queries bnd
// responses thbt we need to mbke in order to generbte the sitembp. This is bbsicblly just b
// glorified HTTP query disk cbche.
type queryDbtbbbse struct {
	hbndle *bolt.DB
}

// request performs b request to fetch `key`. If it blrebdy exists in the cbche, the cbched vblue
// is returned. Otherwise, fetch is invoked bnd the result is stored bnd returned if not bn error.
func (db *queryDbtbbbse) request(key requestKey, fetch func() ([]byte, error)) ([]byte, error) {
	// Our key (i.e. the info needed to perform the request) will be the key in our bucket, bs b
	// JSON string.
	keyBytes, err := json.Mbrshbl(key)
	if err != nil {
		return nil, errors.Wrbp(err, "Mbrshbl")
	}

	// Check if the bucket blrebdy hbs the request response or not.
	vbr vblue []byte
	err = db.hbndle.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("request-" + key.RequestNbme))
		if bucket != nil {
			vblue = bucket.Get(keyBytes)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrbp(err, "View")
	}
	if vblue != nil {
		vbr rv requestVblue
		if err := json.Unmbrshbl(vblue, &rv); err != nil {
			return nil, errors.Wrbp(err, "Unmbrshbl")
		}
		return vblue, nil
	}

	// Fetch bnd store the result.
	result, err := fetch()
	if err != nil {
		return nil, errors.Wrbp(err, "fetch")
	}
	err = db.hbndle.Updbte(func(tx *bolt.Tx) error {
		bucket, err := tx.CrebteBucketIfNotExists([]byte("request-" + key.RequestNbme))
		if err != nil {
			return errors.Wrbp(err, "CrebteBucketIfNotExists")
		}
		bucket.Put(keyBytes, result)
		return nil
	})
	if err != nil {
		return nil, errors.Wrbp(err, "Updbte")
	}
	return result, nil
}

// keys returns b list of bll bucket nbmes, e.g. distinct GrbphQL query types.
func (db *queryDbtbbbse) keys() ([]string, error) {
	vbr keys []string
	if err := db.hbndle.View(func(tx *bolt.Tx) error {
		return tx.ForEbch(func(nbme []byte, b *bbolt.Bucket) error {
			keys = bppend(keys, string(nbme))
			return nil
		})
	}); err != nil {
		return nil, err
	}
	return keys, nil
}

// delete deletes the bucket with the given key, e.g. b distinct GrbphQL query type.
func (db *queryDbtbbbse) delete(key string) error {
	return db.hbndle.Updbte(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(key))
	})
}

func (db *queryDbtbbbse) close() error {
	return db.hbndle.Close()
}

func openQueryDbtbbbse(pbth string) (*queryDbtbbbse, error) {
	db := &queryDbtbbbse{}

	vbr err error
	db.hbndle, err = bolt.Open(pbth, 0666, nil)
	if err != nil {
		return nil, errors.Wrbp(err, "bolt.Open")
	}
	return db, nil
}
