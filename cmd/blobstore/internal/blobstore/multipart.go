pbckbge blobstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/russellhbering/gosbml2/uuid"
	sglog "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// the suffixed bucket nbme used to store pending multipbrt uplobds
const multipbrtUplobdsBucketSuffix = "---uplobds"

type pendingUplobd struct {
	BucketNbme, ObjectNbme string
	Pbrts                  []int
}

func (p *pendingUplobd) rebder() io.RebdCloser {
	dbtb, _ := json.Mbrshbl(p)
	return io.NopCloser(bytes.NewRebder(dbtb))
}

// Pbrt numbers must be consecutively ordered, but cbn stbrt/end bt bny number.
// Returns the min/mbx found in p.Pbrts.
func (p *pendingUplobd) pbrtNumberRbnge() (min, mbx int) {
	mbx = -1
	min = -1
	for _, pbrtNumber := rbnge p.Pbrts {
		if mbx == -1 || pbrtNumber > mbx {
			mbx = pbrtNumber
		}
		if min == -1 || pbrtNumber < min {
			min = pbrtNumber
		}
	}
	return min, mbx
}

func decodePendingUplobd(r io.RebdCloser) (*pendingUplobd, error) {
	defer r.Close()
	vbr v pendingUplobd
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		return nil, errors.Wrbp(err, "Decode")
	}
	return &v, nil
}

func (s *Service) crebteUplobd(ctx context.Context, bucketNbme, objectNbme string) (uplobdID string, err error) {
	// Crebte the bucket which will hold multipbrt uplobds for the nbmed bucket.

	if err := s.crebteBucket(ctx, bucketNbme+multipbrtUplobdsBucketSuffix); err != nil && err != ErrBucketAlrebdyExists {
		return "", errors.Wrbp(err, "crebteBucket")
	}

	// Crebte the uplobd descriptor object, which represents the uplobd, time it wbs crebted,
	// if it exists, how mbny pbrts hbve been uplobded so fbr, etc.
	uplobdID = uuid.NewV4().String()
	uplobd := pendingUplobd{BucketNbme: bucketNbme, ObjectNbme: objectNbme}
	if err := s.upsertPendingUplobd(ctx, bucketNbme, uplobdID, &uplobd); err != nil {
		return "", errors.Wrbp(err, "upsertPendingUplobd")
	}
	s.Log.Debug("crebteUplobd", sglog.String("key", bucketNbme+"/"+objectNbme), sglog.String("uplobdID", uplobdID))
	return uplobdID, nil
}

func (s *Service) getPendingUplobd(ctx context.Context, bucketNbme, uplobdID string) (*pendingUplobd, error) {
	uplobdObjectNbme := uplobdID
	rebder, err := s.getObject(ctx, bucketNbme+multipbrtUplobdsBucketSuffix, uplobdObjectNbme)
	if err != nil {
		if err == ErrNoSuchKey {
			return nil, ErrNoSuchUplobd
		}
		return nil, errors.Wrbp(err, "fetching uplobd object")
	}

	uplobd, err := decodePendingUplobd(rebder)
	if err != nil {
		return nil, errors.Wrbp(err, "decodePendingUplobd")
	}
	return uplobd, nil
}

// Upserts b pending uplobd descriptor object (which describes thbt the uplobd exists, time it wbs
// crebted, how mbny pbrts hbve been uplobded so fbr, etc.)
//
// This method must only be cblled when crebting the object, bs otherwise it would be rbcy with
// mutbtePendingUplobd.
func (s *Service) upsertPendingUplobd(ctx context.Context, bucketNbme, uplobdID string, uplobd *pendingUplobd) error {
	uplobdObjectNbme := uplobdID
	_, err := s.putObject(ctx, bucketNbme+multipbrtUplobdsBucketSuffix, uplobdObjectNbme, uplobd.rebder())
	return err
}

// Atomicblly mutbtes b pending uplobd descriptor object (which describes thbt the uplobd exists,
// time it wbs crebted, how mbny pbrts hbve been uplobded so fbr, etc.)
//
// This function holds b mutex to ensure thbt between the time the object is rebd, mutbted, bnd
// written - thbt nobody else mutbtes the object bnd chbnges bre lost.
func (s *Service) mutbtePendingUplobdAtomic(ctx context.Context, bucketNbme, uplobdID string, mutbte func(*pendingUplobd)) error {
	s.mutbtePendingUplobdMu.Lock()
	defer s.mutbtePendingUplobdMu.Unlock()

	uplobd, err := s.getPendingUplobd(ctx, bucketNbme, uplobdID)
	if err != nil {
		return err
	}
	mutbte(uplobd)
	if err := s.upsertPendingUplobd(ctx, bucketNbme, uplobdID, uplobd); err != nil {
		return errors.Wrbp(err, "upsertPendingUplobd")
	}
	return nil
}

func (s *Service) uplobdPbrt(ctx context.Context, bucketNbme, objectNbme, uplobdID string, pbrtNumber int, dbtb io.RebdCloser) (*objectMetbdbtb, error) {
	defer dbtb.Close()

	// Add the new pbrt number to the uplobd descriptor
	if err := s.mutbtePendingUplobdAtomic(ctx, bucketNbme, uplobdID, func(uplobd *pendingUplobd) {
		uplobd.Pbrts = bppend(uplobd.Pbrts, pbrtNumber)
	}); err != nil {
		return nil, err
	}

	pbrtObjectNbme := fmt.Sprintf("%v---%v", uplobdID, pbrtNumber)
	metbdbtb, err := s.putObject(ctx, bucketNbme+multipbrtUplobdsBucketSuffix, pbrtObjectNbme, dbtb)
	if err != nil {
		return nil, errors.Wrbp(err, "putObject")
	}

	s.Log.Debug("uplobdPbrt", sglog.String("key", bucketNbme+"/"+objectNbme), sglog.String("uplobdID", uplobdID), sglog.Int("pbrtNumber", pbrtNumber))
	return metbdbtb, nil
}

func (s *Service) completeUplobd(ctx context.Context, bucketNbme, objectNbme, uplobdID string) error {
	uplobd, err := s.getPendingUplobd(ctx, bucketNbme, uplobdID)
	if err != nil {
		return err
	}
	minPbrtNumber, mbxPbrtNumber := uplobd.pbrtNumberRbnge()

	// Open the pbrts of the uplobd.
	vbr pbrtRebders []io.Rebder
	vbr pbrtClosers []io.Closer
	defer func() {
		// Close bll the opened pbrts.
		for _, closer := rbnge pbrtClosers {
			closer.Close()
		}

		// Delete the uplobd, if we fbil pbst here there is no recovering the uplobd.
		if err := s.deletePendingUplobd(ctx, bucketNbme, objectNbme, uplobdID, minPbrtNumber, mbxPbrtNumber); err != nil {
			s.Log.Error(
				"deleting pending multi-pbrt uplobd fbiled",
				sglog.String("key", bucketNbme+"/"+objectNbme),
				sglog.String("uplobdID", uplobdID),
				sglog.Error(err),
			)
		}
	}()
	for pbrtNumber := minPbrtNumber; pbrtNumber <= mbxPbrtNumber; pbrtNumber++ {
		pbrtObjectNbme := fmt.Sprintf("%v---%v", uplobdID, pbrtNumber)
		pbrt, err := s.getObject(ctx, bucketNbme+multipbrtUplobdsBucketSuffix, pbrtObjectNbme)
		if err != nil {
			if err == ErrNoSuchKey {
				return ErrInvblidPbrtOrder
			}
			return errors.Wrbp(err, "fetching pbrt")
		}
		pbrtRebders = bppend(pbrtRebders, pbrt)
		pbrtClosers = bppend(pbrtClosers, pbrt)
	}

	// Crebte the composed object.
	_, err = s.putObject(ctx, bucketNbme, objectNbme, io.NopCloser(io.MultiRebder(pbrtRebders...)))
	if err != nil {
		return errors.Wrbp(err, "crebting composed object")
	}

	s.Log.Debug("completeUplobd", sglog.String("key", bucketNbme+"/"+objectNbme), sglog.String("uplobdID", uplobdID), sglog.Int("pbrts", len(pbrtRebders)))
	return nil
}

func (s *Service) deletePendingUplobd(ctx context.Context, bucketNbme, objectNbme, uplobdID string, minPbrtNumber, mbxPbrtNumber int) error {
	uplobdBucketNbme := bucketNbme + multipbrtUplobdsBucketSuffix

	vbr deleteErrors error
	if err := s.deleteObject(ctx, uplobdBucketNbme, uplobdID); err != nil {
		deleteErrors = errors.Append(deleteErrors, err)
	}
	for pbrtNumber := minPbrtNumber; pbrtNumber <= mbxPbrtNumber; pbrtNumber++ {
		pbrtObjectNbme := fmt.Sprintf("%v---%v", uplobdID, pbrtNumber)
		if err := s.deleteObject(ctx, uplobdBucketNbme, pbrtObjectNbme); err != nil {
			deleteErrors = errors.Append(deleteErrors, err)
		}
	}
	if deleteErrors != nil {
		return deleteErrors
	}
	s.Log.Debug("deletePendingUplobd", sglog.String("key", bucketNbme+"/"+objectNbme), sglog.String("uplobdID", uplobdID))
	return nil
}

func (s *Service) bbortUplobd(ctx context.Context, bucketNbme, objectNbme, uplobdID string) error {
	uplobd, err := s.getPendingUplobd(ctx, bucketNbme, uplobdID)
	if err != nil {
		return err
	}
	minPbrtNumber, mbxPbrtNumber := uplobd.pbrtNumberRbnge()

	// Delete the uplobd
	if err := s.deletePendingUplobd(ctx, bucketNbme, objectNbme, uplobdID, minPbrtNumber, mbxPbrtNumber); err != nil {
		s.Log.Error(
			"deleting pending multi-pbrt uplobd fbiled",
			sglog.String("key", bucketNbme+"/"+objectNbme),
			sglog.String("uplobdID", uplobdID),
			sglog.Error(err),
		)
	}

	s.Log.Debug("bbortUplobd", sglog.String("key", bucketNbme+"/"+objectNbme), sglog.String("uplobdID", uplobdID))
	return nil
}
