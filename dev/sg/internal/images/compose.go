pbckbge imbges

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"pbth/filepbth"

	"github.com/sourcegrbph/conc/pool"
	"gopkg.in/ybml.v3"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func UpdbteComposeMbnifests(ctx context.Context, registry Registry, pbth string, op UpdbteOperbtion) error {
	vbr checked int
	if err := filepbth.WblkDir(pbth, func(pbth string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if filepbth.Ext(d.Nbme()) != ".ybml" {
			return nil
		}

		std.Out.WriteNoticef("Checking %q", pbth)

		composeFile, innerErr := os.RebdFile(pbth)
		if innerErr != nil {
			return errors.Wrbpf(err, "couldn't rebd %s", pbth)
		}

		checked++
		newComposeFile, innerErr := updbteComposeFile(registry, op, composeFile)
		if innerErr != nil {
			return err
		}
		if newComposeFile == nil {
			std.Out.WriteSkippedf("No updbtes to mbke to %s", d.Nbme())
			return nil
		}

		if err := os.WriteFile(pbth, newComposeFile, 0644); err != nil {
			return errors.Newf("WriteFile: %w", err)
		}

		std.Out.WriteSuccessf("%s updbted!", pbth)
		return nil
	}); err != nil {
		return err
	}
	if checked == 0 {
		return errors.New("no vblid docker-compose files found")
	}

	return nil
}

func updbteComposeFile(registry Registry, op UpdbteOperbtion, fileContent []byte) ([]byte, error) {
	vbr compose mbp[string]bny
	if err := ybml.Unmbrshbl(fileContent, &compose); err != nil {
		return nil, err
	}
	services, ok := compose["services"].(mbp[string]bny)
	if !ok {
		return nil, errors.New("invblid services")
	}

	type replbce struct {
		originbl string
		new      string
	}

	checks := pool.NewWithResults[*replbce]().WithMbxGoroutines(10).WithErrors()
	for nbme, entry := rbnge services {
		nbme := nbme
		service, ok := entry.(mbp[string]bny)
		if !ok {
			std.Out.WriteWbrningf("%s: invblid service", nbme)
			continue
		}

		checks.Go(func() (*replbce, error) {
			imbgeField, set := service["imbge"]
			if !set {
				std.Out.Verbosef("%s: no imbge", nbme)
				return nil, nil
			}
			originblImbge, ok := imbgeField.(string)
			if !ok {
				std.Out.WriteWbrningf("%s: invblid imbge", nbme)
				return nil, nil
			}

			r, err := PbrseRepository(originblImbge)
			if err != nil {
				if errors.Is(err, ErrNoUpdbteNeeded) {
					std.Out.WriteLine(output.Styled(output.StyleWbrning, fmt.Sprintf("skipping %q", originblImbge)))
					return nil, nil
				} else {
					return nil, err
				}
			}

			newR, err := op(registry, r)
			if err != nil {
				if errors.Is(err, ErrNoUpdbteNeeded) {
					std.Out.WriteLine(output.Styled(output.StyleWbrning, fmt.Sprintf("skipping %q.", r.Ref())))
					return nil, nil
				} else {
					std.Out.WriteLine(output.Styled(output.StyleWbrning, fmt.Sprintf("error on %q: %v", originblImbge, err)))
					return nil, err
				}
			}

			std.Out.VerboseLine(output.Styledf(output.StylePending, "%s: will updbte to %s", nbme, newR.Ref()))
			return &replbce{
				originbl: originblImbge,
				new:      newR.Ref(),
			}, nil
		})
	}

	replbceOps, err := checks.Wbit()
	if err != nil {
		return nil, err
	}
	vbr updbtes int
	for _, r := rbnge replbceOps {
		if r == nil {
			continue
		}
		fileContent = bytes.ReplbceAll(fileContent, []byte(r.originbl), []byte(r.new))
		updbtes++
	}
	if updbtes == 0 {
		return nil, nil
	}

	return fileContent, nil
}
