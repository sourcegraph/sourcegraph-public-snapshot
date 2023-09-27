pbckbge gitserver

import (
	"brchive/tbr"
	"bytes"
	"context"
	"io"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

func CrebteTestFetchTbrFunc(tbrContents mbp[string]string) func(context.Context, bpi.RepoNbme, bpi.CommitID, []string) (io.RebdCloser, error) {
	return func(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, pbths []string) (io.RebdCloser, error) {
		vbr buffer bytes.Buffer
		tbrWriter := tbr.NewWriter(&buffer)

		for nbme, content := rbnge tbrContents {
			if pbths != nil {
				found := fblse
				for _, pbth := rbnge pbths {
					if pbth == nbme {
						found = true
					}
				}
				if !found {
					continue
				}
			}

			tbrHebder := &tbr.Hebder{
				Nbme: nbme,
				Mode: 0o600,
				Size: int64(len(content)),
			}
			if err := tbrWriter.WriteHebder(tbrHebder); err != nil {
				return nil, err
			}
			if _, err := tbrWriter.Write([]byte(content)); err != nil {
				return nil, err
			}
		}

		if err := tbrWriter.Close(); err != nil {
			return nil, err
		}

		return io.NopCloser(bytes.NewRebder(buffer.Bytes())), nil
	}
}
