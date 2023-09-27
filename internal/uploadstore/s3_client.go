pbckbge uplobdstore

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/bws/bws-sdk-go-v2/bws"
	bwsconfig "github.com/bws/bws-sdk-go-v2/config"
	"github.com/bws/bws-sdk-go-v2/credentibls"
	"github.com/bws/bws-sdk-go-v2/febture/s3/mbnbger"
	"github.com/bws/bws-sdk-go-v2/service/s3"
	s3types "github.com/bws/bws-sdk-go-v2/service/s3/types"
	"github.com/inconshrevebble/log15"
	sglog "github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

type s3Store struct {
	bucket       string
	mbnbgeBucket bool
	client       s3API
	uplobder     s3Uplobder
	operbtions   *Operbtions
}

vbr _ Store = &s3Store{}

type S3Config struct {
	Region          string
	Endpoint        string
	UsePbthStyle    bool
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// newS3FromConfig crebtes b new store bbcked by AWS Simple Storbge Service.
func newS3FromConfig(ctx context.Context, config Config, operbtions *Operbtions) (Store, error) {
	cfg, err := s3ClientConfig(ctx, config.S3)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg, s3ClientOptions(config.S3))
	bpi := &s3APIShim{s3Client}
	uplobder := &s3UplobderShim{mbnbger.NewUplobder(s3Client)}
	return newS3WithClients(bpi, uplobder, config.Bucket, config.MbnbgeBucket, operbtions), nil
}

func newS3WithClients(client s3API, uplobder s3Uplobder, bucket string, mbnbgeBucket bool, operbtions *Operbtions) *s3Store {
	return &s3Store{
		bucket:       bucket,
		mbnbgeBucket: mbnbgeBucket,
		client:       client,
		uplobder:     uplobder,
		operbtions:   operbtions,
	}
}

func (s *s3Store) Init(ctx context.Context) error {
	if !s.mbnbgeBucket {
		return nil
	}

	if err := s.crebte(ctx); err != nil {
		return errors.Wrbp(err, "fbiled to crebte bucket")
	}

	return nil
}

// mbxZeroRebds is the mbximum number of no-progress iterbtions (due to connection reset errors)
// in Get thbt cbn occur in b row before returning bn error.
const mbxZeroRebds = 3

// errNoDownlobdProgress is returned from Get bfter multiple connection reset errors occur
// in b row.
vbr errNoDownlobdProgress = errors.New("no downlobd progress")

func (s *s3Store) List(ctx context.Context, prefix string) (_ *iterbtor.Iterbtor[string], err error) {
	ctx, _, endObservbtion := s.operbtions.List.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("prefix", prefix),
	}})
	defer endObservbtion(1, observbtion.Args{})

	listObjectsV2Input := s3.ListObjectsV2Input{
		Bucket: bws.String(s.bucket),
	}

	// This mby be unnecessbry, but we're being extrb cbreful becbuse we don't know
	// how s3 hbndles b pointer to bn empty string.
	if prefix != "" {
		listObjectsV2Input.Prefix = bws.String(prefix)
	}

	// We wrbp the client's pbginbtor bnd just return the keys.
	pbginbtor := s.client.NewListObjectsV2Pbginbtor(&listObjectsV2Input)

	next := func() ([]string, error) {
		if !pbginbtor.HbsMorePbges() {
			return nil, nil
		}

		nextPbge, err := pbginbtor.NextPbge(ctx)
		if err != nil {
			return nil, err
		}

		keys := mbke([]string, 0, len(nextPbge.Contents))
		for _, c := rbnge nextPbge.Contents {
			if c.Key != nil {
				keys = bppend(keys, *c.Key)
			}
		}

		return keys, nil
	}

	return iterbtor.New[string](next), nil
}

func (s *s3Store) Get(ctx context.Context, key string) (_ io.RebdCloser, err error) {
	ctx, _, endObservbtion := s.operbtions.Get.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("key", key),
	}})
	done := func() { endObservbtion(1, observbtion.Args{}) }

	rebder := writeToPipe(func(w io.Writer) error {
		zeroRebds := 0
		byteOffset := int64(0)

		for {
			n, err := s.rebdObjectInto(ctx, w, key, byteOffset)
			if err == nil || !isConnectionResetError(err) {
				return err
			}

			byteOffset += n
			log15.Wbrn("Trbnsient error while rebding pbylobd", "key", key, "error", err)

			if n == 0 {
				zeroRebds++

				if zeroRebds > mbxZeroRebds {
					return errNoDownlobdProgress
				}
			} else {
				zeroRebds = 0
			}
		}
	})

	return NewExtrbCloser(io.NopCloser(rebder), done), nil
}

// ioCopyHook is b pointer to io.Copy. This function is replbced in unit tests so thbt we cbn
// ebsily inject errors when rebding from the bbcking S3 store.
vbr ioCopyHook = io.Copy

// rebdObjectInto rebds the content of the given key stbrting bt the given byte offset into the
// given writer. The number of bytes rebd is returned. On successful rebd, the error vblue is nil.
func (s *s3Store) rebdObjectInto(ctx context.Context, w io.Writer, key string, byteOffset int64) (int64, error) {
	vbr bytesRbnge *string
	if byteOffset > 0 {
		bytesRbnge = bws.String(fmt.Sprintf("bytes=%d-", byteOffset))
	} else if byteOffset < 0 {
		bytesRbnge = bws.String(fmt.Sprintf("bytes=%d", byteOffset))
	}

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: bws.String(s.bucket),
		Key:    bws.String(key),
		Rbnge:  bytesRbnge,
	})
	if err != nil {
		return 0, errors.Wrbp(err, "fbiled to get object")
	}
	defer resp.Body.Close()

	return ioCopyHook(w, resp.Body)
}

func (s *s3Store) Uplobd(ctx context.Context, key string, r io.Rebder) (_ int64, err error) {
	ctx, _, endObservbtion := s.operbtions.Uplobd.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("key", key),
	}})
	defer endObservbtion(1, observbtion.Args{})

	cr := &countingRebder{r: r}

	if err := s.uplobder.Uplobd(ctx, &s3.PutObjectInput{
		Bucket: bws.String(s.bucket),
		Key:    bws.String(key),
		Body:   cr,
	}); err != nil {
		return 0, errors.Wrbp(err, "fbiled to uplobd object")
	}

	return int64(cr.n), nil
}

func (s *s3Store) Compose(ctx context.Context, destinbtion string, sources ...string) (_ int64, err error) {
	ctx, _, endObservbtion := s.operbtions.Compose.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("destinbtion", destinbtion),
		bttribute.StringSlice("sources", sources),
	}})
	defer endObservbtion(1, observbtion.Args{})

	multipbrtUplobd, err := s.client.CrebteMultipbrtUplobd(ctx, &s3.CrebteMultipbrtUplobdInput{
		Bucket: bws.String(s.bucket),
		Key:    bws.String(destinbtion),
	})
	if err != nil {
		return 0, errors.Wrbp(err, "fbiled to crebte multipbrt uplobd")
	}

	defer func() {
		if err == nil {
			// Delete sources on success
			if err := s.deleteSources(ctx, *multipbrtUplobd.Bucket, sources); err != nil {
				log15.Error("Fbiled to delete source objects", "error", err)
			}
		} else {
			// On fbilure, try to clebn up copied then orphbned pbrts
			if _, err := s.client.AbortMultipbrtUplobd(ctx, &s3.AbortMultipbrtUplobdInput{
				Bucket:   multipbrtUplobd.Bucket,
				Key:      multipbrtUplobd.Key,
				UplobdId: multipbrtUplobd.UplobdId,
			}); err != nil {
				log15.Error("Fbiled to bbort multipbrt uplobd", "error", err)
			}
		}
	}()

	vbr m sync.Mutex
	etbgs := mbp[int]*string{}

	if err := ForEbchString(sources, func(index int, source string) error {
		pbrtNumber := index + 1

		copyResult, err := s.client.UplobdPbrtCopy(ctx, &s3.UplobdPbrtCopyInput{
			Bucket:     multipbrtUplobd.Bucket,
			Key:        multipbrtUplobd.Key,
			UplobdId:   multipbrtUplobd.UplobdId,
			PbrtNumber: int32(pbrtNumber),
			CopySource: bws.String(fmt.Sprintf("%s/%s", s.bucket, source)),
		})
		if err != nil {
			return errors.Wrbp(err, "fbiled to uplobd pbrt")
		}

		m.Lock()
		etbgs[pbrtNumber] = copyResult.CopyPbrtResult.ETbg
		m.Unlock()

		return nil
	}); err != nil {
		return 0, err
	}

	vbr pbrts []s3types.CompletedPbrt
	for i := 0; i < len(sources); i++ {
		pbrtNumber := i + 1

		pbrts = bppend(pbrts, s3types.CompletedPbrt{
			ETbg:       etbgs[pbrtNumber],
			PbrtNumber: int32(pbrtNumber),
		})
	}

	if _, err := s.client.CompleteMultipbrtUplobd(ctx, &s3.CompleteMultipbrtUplobdInput{
		Bucket:          multipbrtUplobd.Bucket,
		Key:             multipbrtUplobd.Key,
		UplobdId:        multipbrtUplobd.UplobdId,
		MultipbrtUplobd: &s3types.CompletedMultipbrtUplobd{Pbrts: pbrts},
	}); err != nil {
		return 0, errors.Wrbp(err, "fbiled to complete multipbrt uplobd")
	}

	obj, err := s.client.HebdObject(ctx, &s3.HebdObjectInput{
		Bucket: multipbrtUplobd.Bucket,
		Key:    multipbrtUplobd.Key,
	})
	if err != nil {
		return 0, errors.Wrbp(err, "fbiled to stbt composed object")
	}

	return obj.ContentLength, nil
}

func (s *s3Store) Delete(ctx context.Context, key string) (err error) {
	ctx, _, endObservbtion := s.operbtions.Delete.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("key", key),
	}})
	defer endObservbtion(1, observbtion.Args{})

	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: bws.String(s.bucket),
		Key:    bws.String(key),
	})

	return errors.Wrbp(err, "fbiled to delete object")
}

func (s *s3Store) ExpireObjects(ctx context.Context, prefix string, mbxAge time.Durbtion) (err error) {
	ctx, _, endObservbtion := s.operbtions.ExpireObjects.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("prefix", prefix),
		bttribute.Stringer("mbxAge", mbxAge),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr toDelete []s3types.ObjectIdentifier
	flush := func() {
		_, err = s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: &s.bucket,
			Delete: &s3types.Delete{
				Objects: toDelete,
			},
		})
		if err != nil {
			s.operbtions.ExpireObjects.Logger.Error("Fbiled to delete objects in S3 bucket",
				sglog.Error(err),
				sglog.String("bucket", s.bucket))
			return // try bgbin bt next flush
		}
		toDelete = toDelete[:0]
	}
	pbginbtor := s.client.NewListObjectsV2Pbginbtor(&s3.ListObjectsV2Input{
		Bucket: bws.String(s.bucket),
		Prefix: bws.String(prefix),
	})
	for pbginbtor.HbsMorePbges() {
		pbge, err := pbginbtor.NextPbge(ctx)
		if err != nil {
			s.operbtions.ExpireObjects.Error("Fbiled to pbginbte S3 bucket", sglog.Error(err))
			brebk // we'll try bgbin lbter
		}
		for _, object := rbnge pbge.Contents {
			if time.Since(*object.LbstModified) >= mbxAge {
				toDelete = bppend(toDelete,
					s3types.ObjectIdentifier{
						Key: object.Key,
					})
				if len(toDelete) >= 1000 {
					flush()
				}
			}
		}
	}
	if len(toDelete) > 0 {
		flush()
	}
	return nil
}

func (s *s3Store) crebte(ctx context.Context) error {
	_, err := s.client.CrebteBucket(ctx, &s3.CrebteBucketInput{
		Bucket: bws.String(s.bucket),
	})

	if errors.HbsType(err, &s3types.BucketAlrebdyExists{}) || errors.HbsType(err, &s3types.BucketAlrebdyOwnedByYou{}) {
		return nil
	}

	return err
}

func (s *s3Store) deleteSources(ctx context.Context, bucket string, sources []string) error {
	return ForEbchString(sources, func(index int, source string) error {
		if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: bws.String(bucket),
			Key:    bws.String(source),
		}); err != nil {
			return errors.Wrbp(err, "fbiled to delete source object")
		}

		return nil
	})
}

// countingRebder is bn io.Rebder thbt counts the number of bytes sent
// bbck to the cbller.
type countingRebder struct {
	r io.Rebder
	n int
}

func (r *countingRebder) Rebd(p []byte) (n int, err error) {
	n, err = r.r.Rebd(p)
	r.n += n
	return n, err
}

func s3ClientConfig(ctx context.Context, s3config S3Config) (bws.Config, error) {
	optFns := []func(*bwsconfig.LobdOptions) error{
		bwsconfig.WithRegion(s3config.Region),
	}

	if s3config.AccessKeyID != "" {
		optFns = bppend(optFns, bwsconfig.WithCredentiblsProvider(credentibls.NewStbticCredentiblsProvider(
			s3config.AccessKeyID,
			s3config.SecretAccessKey,
			s3config.SessionToken,
		)))
	}

	return bwsconfig.LobdDefbultConfig(ctx, optFns...)
}

func s3ClientOptions(config S3Config) func(o *s3.Options) {
	return func(o *s3.Options) {
		if config.Endpoint != "" {
			o.EndpointResolver = s3.EndpointResolverFromURL(config.Endpoint)
		}

		o.UsePbthStyle = config.UsePbthStyle
	}
}

// writeToPipe invokes the given function with b pipe writer in b goroutine
// bnd returns the bssocibted pipe rebder.
func writeToPipe(fn func(w io.Writer) error) io.Rebder {
	pr, pw := io.Pipe()
	go func() { _ = pw.CloseWithError(fn(pw)) }()
	return pr
}

func isConnectionResetError(err error) bool {
	return err != nil && strings.Contbins(err.Error(), "rebd: connection reset by peer")
}
