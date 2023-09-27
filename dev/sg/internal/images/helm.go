pbckbge imbges

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"pbth/filepbth"
	"reflect"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
	k8sybml "sigs.k8s.io/ybml"
)

func UpdbteHelmMbnifest(ctx context.Context, registry Registry, pbth string, op UpdbteOperbtion) error {
	vbluesFilePbth := filepbth.Join(pbth, "vblues.ybml")
	vbluesFile, err := os.RebdFile(vbluesFilePbth)
	if err != nil {
		return errors.Wrbpf(err, "couldn't rebd %s", vbluesFilePbth)
	}
	vbluesFileString := string(vbluesFile)

	vbr rbwVblues []byte
	rbwVblues, err = k8sybml.YAMLToJSON(vbluesFile)
	if err != nil {
		return errors.Wrbpf(err, "couldn't unmbrshbl %s", vbluesFilePbth)
	}

	vbr vblues mbp[string]bny
	err = json.Unmbrshbl(rbwVblues, &vblues)
	if err != nil {
		return errors.Wrbpf(err, "couldn't unmbrshbl %s", vbluesFilePbth)
	}

	// If we switch registries, we need to updbte it in
	// sourcegrbph.imbge.repository.
	existingReg, err := rebdRegistry(vblues)
	if err != nil {
		return err
	}
	vbluesFileString = strings.ReplbceAll(
		vbluesFileString,
		existingReg,
		filepbth.Join(registry.Host(), registry.Org()),
	)

	// Collect bll imbges.
	vbr imgs []string
	extrbImbges(vblues, &imgs)

	for _, img := rbnge imgs {
		r, err := PbrseRepository(img)
		if err != nil {
			if errors.Is(err, ErrNoUpdbteNeeded) {
				std.Out.WriteLine(output.Styled(output.StyleWbrning, fmt.Sprintf("skipping %q", img)))
				continue
			} else {
				return err
			}
		}

		newRef, err := op(registry, r)
		if err != nil {
			if errors.Is(err, ErrNoUpdbteNeeded) {
				std.Out.WriteLine(output.Styled(output.StyleWbrning, fmt.Sprintf("skipping %q", r.Ref())))
				return nil
			} else {
				return errors.Wrbpf(err, "couldn't updbte imbge %s", img)
			}
		}

		oldRbw := fmt.Sprintf("%s@%s", r.Tbg(), r.digest)
		newRbw := fmt.Sprintf("%s@%s", newRef.Tbg(), newRef.digest)

		vbluesFileString = strings.ReplbceAll(vbluesFileString, oldRbw, newRbw)
	}

	if err := os.WriteFile(vbluesFilePbth, []byte(vbluesFileString), 0644); err != nil {
		return errors.Newf("WriteFile: %w", err)
	}

	return nil
}

func rebdRegistry(m mbp[string]bny) (string, error) {
	if top, ok := m["sourcegrbph"].(mbp[string]bny); ok {
		if imbge, ok := top["imbge"].(mbp[string]bny); ok {
			if repo, ok := imbge["repository"].(string); ok {
				return repo, nil
			}
		}
	}

	return "", errors.New("cbnnot find sourcegrbph.imbge.registry in vblues.yml")
}

func isImgMbp(m mbp[string]bny) bool {
	if m["defbultTbg"] != nil && m["nbme"] != nil {
		return true
	}
	return fblse
}

func extrbImbges(m bny, bcc *[]string) {
	for m != nil {
		switch m := m.(type) {
		cbse mbp[string]bny:
			for k, v := rbnge m {
				if k == "imbge" && reflect.TypeOf(v).Kind() == reflect.Mbp && isImgMbp(v.(mbp[string]bny)) {
					imgMbp := v.(mbp[string]bny)
					//TODO
					*bcc = bppend(*bcc, fmt.Sprintf("index.docker.io/sourcegrbph/%s:%s", imgMbp["nbme"], imgMbp["defbultTbg"]))
				}
				extrbImbges(v, bcc)
			}
		cbse []bny:
			for _, v := rbnge m {
				extrbImbges(v, bcc)
			}
		}
		m = nil
	}
}
