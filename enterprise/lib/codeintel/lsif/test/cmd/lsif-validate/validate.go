package main

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/efritz/pentimento"
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/test/cmd/lsif-validate/internal/validation"
)

var updateInterval = time.Second / 4
var ticker = pentimento.NewAnimatedString([]string{"⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏", "⠋", "⠙", "⠹"}, updateInterval)

func validate(indexFile *os.File) error {
	ctx := validation.NewValidationContext()
	validator := &validation.Validator{Context: ctx}
	errs := make(chan error, 1)

	go func() {
		defer close(errs)

		if err := validator.Validate(indexFile); err != nil {
			errs <- err
		}
	}()

	if err := printProgress(ctx, validator, errs); err != nil {
		return err
	}

	for i, err := range ctx.Errors {
		fmt.Printf("%d) %s\n", i+1, err)
	}

	if len(ctx.Errors) > 0 {
		return errors.New(fmt.Sprintf("Detected %d errors", len(ctx.Errors)))
	}

	return nil
}

func printProgress(ctx *validation.ValidationContext, validator *validation.Validator, errs <-chan error) error {
	return pentimento.PrintProgress(func(printer *pentimento.Printer) error {
		defer func() {
			_ = printer.Reset()
		}()

		for {
			ctx.ErrorsLock.RLock()
			numErrors := len(ctx.Errors)
			ctx.ErrorsLock.RUnlock()

			content := pentimento.NewContent()
			content.AddLine(
				"%s %d vertices, %d edges, %d errors",
				ticker,
				atomic.LoadUint64(&ctx.NumVertices),
				atomic.LoadUint64(&ctx.NumEdges),
				numErrors,
			)
			printer.WriteContent(content)

			select {
			case err := <-errs:
				return err
			case <-time.After(updateInterval):
			}
		}
	})
}
