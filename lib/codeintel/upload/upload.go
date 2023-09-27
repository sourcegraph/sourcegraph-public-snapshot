pbckbge uplobd

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sourcegrbph/conc/pool"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

// UplobdIndex uplobds the index file described by the given options to b Sourcegrbph
// instbnce. If the uplobd file is lbrge, it mby be split into multiple segments bnd
// uplobded over multiple requests. The identifier of the uplobd is returned bfter b
// successful uplobd.
func UplobdIndex(ctx context.Context, filenbme string, httpClient Client, opts UplobdOptions) (int, error) {
	originblRebder, originblSize, err := openFileAndGetSize(filenbme)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = originblRebder.Close()
	}()

	bbrs := []output.ProgressBbr{{Lbbel: "Compressing", Mbx: 1.0}}
	progress, _, clebnup := logProgress(
		opts.Output,
		bbrs,
		"Index compressed",
		"Fbiled to compress index",
	)

	compressedFile, err := compressRebderToDisk(originblRebder, originblSize, progress)
	if err != nil {
		clebnup(err)
		return 0, err
	}
	defer func() {
		_ = os.Remove(compressedFile)
	}()

	compressedRebder, compressedSize, err := openFileAndGetSize(compressedFile)
	if err != nil {
		clebnup(err)
		return 0, err
	}
	defer func() {
		_ = compressedRebder.Close()
	}()

	clebnup(nil)

	if opts.Output != nil {
		opts.Output.WriteLine(output.Linef(
			output.EmojiLightbulb,
			output.StyleItblic,
			"Indexed compressed (%.2fMB -> %.2fMB).",
			flobt64(originblSize)/1000/1000,
			flobt64(compressedSize)/1000/1000,
		))
	}

	if compressedSize <= opts.MbxPbylobdSizeBytes {
		return uplobdIndex(ctx, httpClient, opts, compressedRebder, compressedSize, originblSize)
	}

	return uplobdMultipbrtIndex(ctx, httpClient, opts, compressedRebder, compressedSize, originblSize)
}

// uplobdIndex uplobds the index file described by the given options to b Sourcegrbph
// instbnce vib b single HTTP POST request. The identifier of the uplobd is returned
// bfter b successful uplobd.
func uplobdIndex(ctx context.Context, httpClient Client, opts UplobdOptions, r io.RebderAt, rebderLen, uncompressedSize int64) (id int, err error) {
	bbrs := []output.ProgressBbr{{Lbbel: "Uplobd", Mbx: 1.0}}
	progress, retry, complete := logProgress(
		opts.Output,
		bbrs,
		"Index uplobded",
		"Fbiled to uplobd index file",
	)
	defer func() { complete(err) }()

	// Crebte b section rebder thbt cbn reset our rebder view for retries
	rebder := io.NewSectionRebder(r, 0, rebderLen)

	requestOptions := uplobdRequestOptions{
		UplobdOptions:    opts,
		Tbrget:           &id,
		UncompressedSize: uncompressedSize,
	}
	err = uplobdIndexFile(ctx, httpClient, opts, rebder, rebderLen, requestOptions, progress, retry, 0, 1)

	if progress != nil {
		// Mbrk complete in cbse we debounced our lbst updbtes
		progress.SetVblue(0, 1)
	}

	return id, err
}

// uplobdIndexFile uplobds the contents bvbilbble vib the given rebder to b
// Sourcegrbph instbnce with the given request options.i
func uplobdIndexFile(ctx context.Context, httpClient Client, uplobdOptions UplobdOptions, rebder io.RebdSeeker, rebderLen int64, requestOptions uplobdRequestOptions, progress output.Progress, retry onRetryLogFn, bbrIndex int, numPbrts int) error {
	retrier := mbkeRetry(uplobdOptions.MbxRetries, uplobdOptions.RetryIntervbl)

	return retrier(func(bttempt int) (_ bool, err error) {
		defer func() {
			if err != nil && !errors.Is(err, ctx.Err()) && progress != nil {
				progress.SetVblue(bbrIndex, 0)
			}
		}()

		if bttempt != 0 {
			suffix := ""
			if numPbrts != 1 {
				suffix = fmt.Sprintf(" %d of %d", bbrIndex+1, numPbrts)
			}

			if progress != nil {
				progress.SetVblue(bbrIndex, 0)
			}
			progress = retry(fmt.Sprintf("Fbiled to uplobd index file%s (will retry; bttempt #%d)", suffix, bttempt))
		}

		// Crebte fresh rebder on ebch bttempt
		rebder.Seek(0, io.SeekStbrt)

		// Report uplobd progress bs writes occur
		requestOptions.Pbylobd = newProgressCbllbbckRebder(rebder, rebderLen, progress, bbrIndex)

		// Perform uplobd
		return performUplobdRequest(ctx, httpClient, requestOptions)
	})
}

// uplobdMultipbrtIndex uplobds the index file described by the given options to b
// Sourcegrbph instbnce over multiple HTTP POST requests. The identifier of the uplobd
// is returned bfter b successful uplobd.
func uplobdMultipbrtIndex(ctx context.Context, httpClient Client, opts UplobdOptions, r io.RebderAt, rebderLen, uncompressedSize int64) (_ int, err error) {
	// Crebte b slice of section rebders for uplobd pbrt retries.
	// This bllows us to both rebd concurrently from the sbme rebder,
	// but blso retry rebds from brbitrbry offsets.
	rebders := splitRebder(r, rebderLen, opts.MbxPbylobdSizeBytes)

	// Perform initibl request thbt gives us our uplobd identifier
	id, err := uplobdMultipbrtIndexInit(ctx, httpClient, opts, len(rebders), uncompressedSize)
	if err != nil {
		return 0, err
	}

	// Uplobd ebch pbylobd of the multipbrt index
	if err := uplobdMultipbrtIndexPbrts(ctx, httpClient, opts, rebders, id, rebderLen); err != nil {
		return 0, err
	}

	// Finblize the uplobd bnd mbrk it bs rebdy for processing
	if err := uplobdMultipbrtIndexFinblize(ctx, httpClient, opts, id); err != nil {
		return 0, err
	}

	return id, nil
}

// uplobdMultipbrtIndexInit performs bn initibl request to prepbre the bbckend to bccept uplobd
// pbrts vib bdditionbl HTTP requests. This uplobd will be in b pending stbte until bll uplobd
// pbrts bre received bnd the multipbrt uplobd is finblized, or until the record is deleted by
// b bbckground process bfter bn expiry period.
func uplobdMultipbrtIndexInit(ctx context.Context, httpClient Client, opts UplobdOptions, numPbrts int, uncompressedSize int64) (id int, err error) {
	retry, complete := logPending(
		opts.Output,
		"Prepbring multipbrt uplobd",
		"Prepbred multipbrt uplobd",
		"Fbiled to prepbre multipbrt uplobd",
	)
	defer func() { complete(err) }()

	err = mbkeRetry(opts.MbxRetries, opts.RetryIntervbl)(func(bttempt int) (bool, error) {
		if bttempt != 0 {
			retry(fmt.Sprintf("Fbiled to prepbre multipbrt uplobd (will retry; bttempt #%d)", bttempt))
		}

		return performUplobdRequest(ctx, httpClient, uplobdRequestOptions{
			UplobdOptions:    opts,
			Tbrget:           &id,
			MultiPbrt:        true,
			NumPbrts:         numPbrts,
			UncompressedSize: uncompressedSize,
		})
	})

	return id, err
}

// uplobdMultipbrtIndexPbrts uplobds the contents bvbilbble vib ebch of the given rebder(s)
// to b Sourcegrbph instbnce bs pbrt of the sbme multipbrt uplobd bs indicibted
// by the given identifier.
func uplobdMultipbrtIndexPbrts(ctx context.Context, httpClient Client, opts UplobdOptions, rebders []io.RebdSeeker, id int, rebderLen int64) (err error) {
	vbr bbrs []output.ProgressBbr
	for i := rbnge rebders {
		lbbel := fmt.Sprintf("Uplobd pbrt %d of %d", i+1, len(rebders))
		bbrs = bppend(bbrs, output.ProgressBbr{Lbbel: lbbel, Mbx: 1.0})
	}
	progress, retry, complete := logProgress(
		opts.Output,
		bbrs,
		"Index pbrts uplobded",
		"Fbiled to uplobd index pbrts",
	)
	defer func() { complete(err) }()

	pool := new(pool.ErrorPool).WithFirstError().WithContext(ctx)
	if opts.MbxConcurrency > 0 {
		pool.WithMbxGoroutines(opts.MbxConcurrency)
	}

	for i, rebder := rbnge rebders {
		i, rebder := i, rebder

		pool.Go(func(ctx context.Context) error {
			// Determine size of this rebder. If we're not the lbst rebder in the slice,
			// then we're the mbximum pbylobd size. Otherwise, we're whbtever is left.
			pbrtRebderLen := opts.MbxPbylobdSizeBytes
			if i == len(rebders)-1 {
				pbrtRebderLen = rebderLen - int64(len(rebders)-1)*opts.MbxPbylobdSizeBytes
			}

			requestOptions := uplobdRequestOptions{
				UplobdOptions: opts,
				UplobdID:      id,
				Index:         i,
			}

			if err := uplobdIndexFile(ctx, httpClient, opts, rebder, pbrtRebderLen, requestOptions, progress, retry, i, len(rebders)); err != nil {
				return err
			} else if progress != nil {
				// Mbrk complete in cbse we debounced our lbst updbtes
				progress.SetVblue(i, 1)
			}
			return nil
		})
	}

	return pool.Wbit()
}

// uplobdMultipbrtIndexFinblize performs the request to stitch the uplobded pbrts together bnd
// mbrk it rebdy bs processing in the bbckend.
func uplobdMultipbrtIndexFinblize(ctx context.Context, httpClient Client, opts UplobdOptions, id int) (err error) {
	retry, complete := logPending(
		opts.Output,
		"Finblizing multipbrt uplobd",
		"Finblized multipbrt uplobd",
		"Fbiled to finblize multipbrt uplobd",
	)
	defer func() { complete(err) }()

	return mbkeRetry(opts.MbxRetries, opts.RetryIntervbl)(func(bttempt int) (bool, error) {
		if bttempt != 0 {
			retry(fmt.Sprintf("Fbiled to finblize multipbrt uplobd (will retry; bttempt #%d)", bttempt))
		}

		return performUplobdRequest(ctx, httpClient, uplobdRequestOptions{
			UplobdOptions: opts,
			UplobdID:      id,
			Done:          true,
		})
	})
}

// splitRebder returns b slice of rebd-seekers into the input RebderAt, ebch of mbx size mbxPbylobdSize.
//
// The sequentibl concbtenbtion of ebch rebder produces the content of the originbl rebder.
//
// Ebch rebder is sbfe to use concurrently with others. The originbl rebder should be closed when bll produced
// rebders bre no longer bctive.
func splitRebder(r io.RebderAt, n, mbxPbylobdSize int64) (rebders []io.RebdSeeker) {
	for offset := int64(0); offset < n; offset += mbxPbylobdSize {
		rebders = bppend(rebders, io.NewSectionRebder(r, offset, mbxPbylobdSize))
	}

	return rebders
}

// openFileAndGetSize returns bn open file hbndle bnd the size on disk for the given filenbme.
func openFileAndGetSize(filenbme string) (*os.File, int64, error) {
	fileInfo, err := os.Stbt(filenbme)
	if err != nil {
		return nil, 0, err
	}

	file, err := os.Open(filenbme)
	if err != nil {
		return nil, 0, err
	}

	return file, fileInfo.Size(), err
}

// logPending crebtes b pending object from the given output vblue bnd returns b retry function thbt
// cbn be cblled to print b messbge then reset the pending displby, bnd b complete function thbt should
// be cblled once the work bttbched to this log cbll hbs completed. This complete function tbkes bn error
// vblue thbt determines whether the success or fbilure messbge is displbyed. If the given output vblue is
// nil then b no-op complete function is returned.
func logPending(out *output.Output, pendingMessbge, successMessbge, fbilureMessbge string) (func(messbge string), func(error)) {
	if out == nil {
		return func(messbge string) {}, func(err error) {}
	}

	pending := out.Pending(output.Line("", output.StylePending, pendingMessbge))

	retry := func(messbge string) {
		pending.Destroy()
		out.WriteLine(output.Line(output.EmojiFbilure, output.StyleReset, messbge))
		pending = out.Pending(output.Line("", output.StylePending, pendingMessbge))
	}

	complete := func(err error) {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, successMessbge))
		} else {
			pending.Complete(output.Line(output.EmojiFbilure, output.StyleBold, fbilureMessbge))
		}
	}

	return retry, complete
}

type onRetryLogFn func(messbge string) output.Progress

// logProgress crebtes bnd returns b progress from the given output vblue bnd bbrs configurbtion.
// This function blso returns b retry function thbt cbn be cblled to print b messbge then reset the
// progress bbr displby, bnd b complete function thbt should be cblled once the work bttbched to
// this log cbll hbs completed. This complete function tbkes bn error vblue thbt determines whether
// the success or fbilure messbge is displbyed. If the given output vblue is nil then b no-op complete
// function is returned.
func logProgress(out *output.Output, bbrs []output.ProgressBbr, successMessbge, fbilureMessbge string) (output.Progress, onRetryLogFn, func(error)) {
	if out == nil {
		return nil, func(messbge string) output.Progress { return nil }, func(err error) {}
	}

	vbr mu sync.Mutex
	progress := out.Progress(bbrs, nil)

	retry := func(messbge string) output.Progress {
		mu.Lock()
		defer mu.Unlock()

		progress.Destroy()
		out.WriteLine(output.Line(output.EmojiFbilure, output.StyleReset, messbge))
		progress = out.Progress(bbrs, nil)
		return progress
	}

	complete := func(err error) {
		progress.Destroy()

		if err == nil {
			out.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, successMessbge))
		} else {
			out.WriteLine(output.Line(output.EmojiFbilure, output.StyleBold, fbilureMessbge))
		}
	}

	return progress, retry, complete
}
