package upload

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// UploadIndex uploads the index file described by the given options to a Sourcegraph
// instance. If the upload file is large, it may be split into multiple segments and
// uploaded over multiple requests. The identifier of the upload is returned after a
// successful upload.
func UploadIndex(ctx context.Context, filename string, httpClient Client, opts UploadOptions) (int, error) {
	originalReader, originalSize, err := openFileAndGetSize(filename)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = originalReader.Close()
	}()

	bars := []output.ProgressBar{{Label: "Compressing", Max: 1.0}}
	progress, _, cleanup := logProgress(
		opts.Output,
		bars,
		"Index compressed",
		"Failed to compress index",
	)

	compressedFile, err := compressReaderToDisk(originalReader, originalSize, progress)
	if err != nil {
		cleanup(err)
		return 0, err
	}
	defer func() {
		_ = os.Remove(compressedFile)
	}()

	compressedReader, compressedSize, err := openFileAndGetSize(compressedFile)
	if err != nil {
		cleanup(err)
		return 0, err
	}
	defer func() {
		_ = compressedReader.Close()
	}()

	cleanup(nil)

	if opts.Output != nil {
		opts.Output.WriteLine(output.Linef(
			output.EmojiLightbulb,
			output.StyleItalic,
			"Indexed compressed (%.2fMB -> %.2fMB).",
			float64(originalSize)/1000/1000,
			float64(compressedSize)/1000/1000,
		))
	}

	if compressedSize <= opts.MaxPayloadSizeBytes {
		return uploadIndex(ctx, httpClient, opts, compressedReader, compressedSize)
	}

	return uploadMultipartIndex(ctx, httpClient, opts, compressedReader, compressedSize)
}

// uploadIndex uploads the index file described by the given options to a Sourcegraph
// instance via a single HTTP POST request. The identifier of the upload is returned
// after a successful upload.
func uploadIndex(ctx context.Context, httpClient Client, opts UploadOptions, r io.ReaderAt, readerLen int64) (id int, err error) {
	bars := []output.ProgressBar{{Label: "Upload", Max: 1.0}}
	progress, retry, complete := logProgress(
		opts.Output,
		bars,
		"Index uploaded",
		"Failed to upload index file",
	)
	defer func() { complete(err) }()

	// Create a function that can re-create our reader for retries
	readerFactory := func() io.Reader { return newBoundedReader(r, 0, readerLen) }

	requestOptions := uploadRequestOptions{
		UploadOptions: opts,
		Target:        &id,
	}
	err = uploadIndexFile(ctx, httpClient, opts, readerFactory, readerLen, requestOptions, progress, retry, 0, 1)

	if progress != nil {
		// Mark complete in case we debounced our last updates
		progress.SetValue(0, 1)
	}

	return id, err
}

// uploadIndexFile uploads the contents available via the given reader factory to a
// Sourcegraph instance with the given request options.i
func uploadIndexFile(ctx context.Context, httpClient Client, uploadOptions UploadOptions, readerFactory func() io.Reader, readerLen int64, requestOptions uploadRequestOptions, progress output.Progress, retry func(message string) output.Progress, barIndex int, numParts int) error {
	return makeRetry(uploadOptions.MaxRetries, uploadOptions.RetryInterval)(func(attempt int) (_ bool, err error) {
		defer func() {
			if err != nil && !errors.Is(err, ctx.Err()) && progress != nil {
				progress.SetValue(barIndex, 0)
			}
		}()

		if attempt != 0 {
			suffix := ""
			if numParts != 1 {
				suffix = fmt.Sprintf(" %d of %d", barIndex+1, numParts)
			}

			if progress != nil {
				progress.SetValue(barIndex, 0)
			}
			progress = retry(fmt.Sprintf("Failed to upload index file%s (will retry; attempt #%d)", suffix, attempt))
		}

		// Create fresh reader on each attempt
		reader := readerFactory()

		// Report upload progress as writes occur
		requestOptions.Payload = newProgressCallbackReader(reader, readerLen, progress, barIndex)

		// Perform upload
		return performUploadRequest(ctx, httpClient, requestOptions)
	})
}

// uploadMultipartIndex uploads the index file described by the given options to a
// Sourcegraph instance over multiple HTTP POST requests. The identifier of the upload
// is returned after a successful upload.
func uploadMultipartIndex(ctx context.Context, httpClient Client, opts UploadOptions, r io.ReaderAt, readerLen int64) (_ int, err error) {
	// Create a slice of functions that can re-create our reader for an
	// upload part for retries. This allows us to both read concurrently
	// from the same reader, but also retry reads from arbitrary offsets.
	readerFactories := splitReader(r, readerLen, opts.MaxPayloadSizeBytes)

	// Perform initial request that gives us our upload identifier
	id, err := uploadMultipartIndexInit(ctx, httpClient, opts, len(readerFactories))
	if err != nil {
		return 0, err
	}

	// Upload each payload of the multipart index
	if err := uploadMultipartIndexParts(ctx, httpClient, opts, readerFactories, id, readerLen); err != nil {
		return 0, err
	}

	// Finalize the upload and mark it as ready for processing
	if err := uploadMultipartIndexFinalize(ctx, httpClient, opts, id); err != nil {
		return 0, err
	}

	return id, nil
}

// uploadMultipartIndexInit performs an initial request to prepare the backend to accept upload
// parts via additional HTTP requests. This upload will be in a pending state until all upload
// parts are received and the multipart upload is finalized, or until the record is deleted by
// a background process after an expiry period.
func uploadMultipartIndexInit(ctx context.Context, httpClient Client, opts UploadOptions, numParts int) (id int, err error) {
	retry, complete := logPending(
		opts.Output,
		"Preparing multipart upload",
		"Prepared multipart upload",
		"Failed to prepare multipart upload",
	)
	defer func() { complete(err) }()

	err = makeRetry(opts.MaxRetries, opts.RetryInterval)(func(attempt int) (bool, error) {
		if attempt != 0 {
			retry(fmt.Sprintf("Failed to prepare multipart upload (will retry; attempt #%d)", attempt))
		}

		return performUploadRequest(ctx, httpClient, uploadRequestOptions{
			UploadOptions: opts,
			Target:        &id,
			MultiPart:     true,
			NumParts:      numParts,
		})
	})

	return id, err
}

// uploadMultipartIndexParts uploads the contents available via each of the given reader
// factories to a Sourcegraph instance as part of the same multipart upload as indiciated
// by the given identifier.
func uploadMultipartIndexParts(ctx context.Context, httpClient Client, opts UploadOptions, readerFactories []func() io.Reader, id int, readerLen int64) (err error) {
	var bars []output.ProgressBar
	for i := range readerFactories {
		label := fmt.Sprintf("Upload part %d of %d", i+1, len(readerFactories))
		bars = append(bars, output.ProgressBar{Label: label, Max: 1.0})
	}
	progress, retry, complete := logProgress(
		opts.Output,
		bars,
		"Index parts uploaded",
		"Failed to upload index parts",
	)
	defer func() { complete(err) }()

	var wg sync.WaitGroup
	errs := make(chan error, len(readerFactories))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, readerFactory := range readerFactories {
		wg.Add(1)

		go func(i int, readerFactory func() io.Reader) {
			defer wg.Done()

			// Determine size of this reader. If we're not the last reader in the slice,
			// then we're the maximum payload size. Otherwise, we're whatever is left.
			partReaderLen := opts.MaxPayloadSizeBytes
			if i == len(readerFactories)-1 {
				partReaderLen = readerLen - int64(len(readerFactories)-1)*opts.MaxPayloadSizeBytes
			}

			requestOptions := uploadRequestOptions{
				UploadOptions: opts,
				UploadID:      id,
				Index:         i,
			}

			if err := uploadIndexFile(ctx, httpClient, opts, readerFactory, partReaderLen, requestOptions, progress, retry, i, len(readerFactories)); err != nil {
				errs <- err
				cancel()
			} else if progress != nil {
				// Mark complete in case we debounced our last updates
				progress.SetValue(i, 1)
			}
		}(i, readerFactory)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

// uploadMultipartIndexFinalize performs the request to stitch the uploaded parts together and
// mark it ready as processing in the backend.
func uploadMultipartIndexFinalize(ctx context.Context, httpClient Client, opts UploadOptions, id int) (err error) {
	retry, complete := logPending(
		opts.Output,
		"Finalizing multipart upload",
		"Finalized multipart upload",
		"Failed to finalize multipart upload",
	)
	defer func() { complete(err) }()

	return makeRetry(opts.MaxRetries, opts.RetryInterval)(func(attempt int) (bool, error) {
		if attempt != 0 {
			retry(fmt.Sprintf("Failed to finalize multipart upload (will retry; attempt #%d)", attempt))
		}

		return performUploadRequest(ctx, httpClient, uploadRequestOptions{
			UploadOptions: opts,
			UploadID:      id,
			Done:          true,
		})
	})
}

// splitReader returns a slice of functions, each which returns a fresh instance of a reader
// looking into a slice of the original reader.
//
// Two readers from the same factory contain the same content that allows retrying reads from
// the beginning. The sequential concatenation of each reader produces the content of the original
// reader.
//
// Each reader is safe to use concurrently. The original reader should be closed when all produced
// readers are no longer active.
func splitReader(r io.ReaderAt, n, maxPayloadSize int64) (readerFactories []func() io.Reader) {
	for i := int64(0); i < n; i += maxPayloadSize {
		minOffset := i
		maxOffset := minOffset + maxPayloadSize
		if maxOffset > n {
			maxOffset = n
		}

		readerFactories = append(readerFactories, func() io.Reader {
			return newBoundedReader(r, minOffset, maxOffset)
		})
	}

	return readerFactories
}

// openFileAndGetSize returns an open file handle and the size on disk for the given filename.
func openFileAndGetSize(filename string) (*os.File, int64, error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return nil, 0, err
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, err
	}

	return file, fileInfo.Size(), err
}

// logPending creates a pending object from the given output value and returns a retry function that
// can be called to print a message then reset the pending display, and a complete function that should
// be called once the work attached to this log call has completed. This complete function takes an error
// value that determines whether the success or failure message is displayed. If the given output value is
// nil then a no-op complete function is returned.
func logPending(out *output.Output, pendingMessage, successMessage, failureMessage string) (func(message string), func(error)) {
	if out == nil {
		return func(message string) {}, func(err error) {}
	}

	pending := out.Pending(output.Line("", output.StylePending, pendingMessage))

	retry := func(message string) {
		pending.Destroy()
		out.WriteLine(output.Line(output.EmojiFailure, output.StyleReset, message))
		pending = out.Pending(output.Line("", output.StylePending, pendingMessage))
	}

	complete := func(err error) {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, successMessage))
		} else {
			pending.Complete(output.Line(output.EmojiFailure, output.StyleBold, failureMessage))
		}
	}

	return retry, complete
}

// logProgress creates and returns a progress from the given output value and bars configuration.
// This function also returns a retry function that can be called to print a message then reset the
// progress bar display, and a complete function that should be called once the work attached to
// this log call has completed. This complete function takes an error value that determines whether
// the success or failure message is displayed. If the given output value is nil then a no-op complete
// function is returned.
func logProgress(out *output.Output, bars []output.ProgressBar, successMessage, failureMessage string) (output.Progress, func(message string) output.Progress, func(error)) {
	if out == nil {
		return nil, func(message string) output.Progress { return nil }, func(err error) {}
	}

	var mu sync.Mutex
	progress := out.Progress(bars, nil)

	retry := func(message string) output.Progress {
		mu.Lock()
		defer mu.Unlock()

		progress.Destroy()
		out.WriteLine(output.Line(output.EmojiFailure, output.StyleReset, message))
		progress = out.Progress(bars, nil)
		return progress
	}

	complete := func(err error) {
		progress.Destroy()

		if err == nil {
			out.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, successMessage))
		} else {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, failureMessage))
		}
	}

	return progress, retry, complete
}
