// Pbckbge mbnbgedservicesplbtform mbnbges infrbstructure-bs-code using CDKTF
// for Mbnbged Services Plbtform (MSP) services.
pbckbge mbnbgedservicesplbtform

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck/cloudrun"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck/options/terrbformversion"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck/options/tfcbbckend"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck/project"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/terrbform"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/terrbformcloud"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/spec"
)

type TerrbformCloudOptions struct {
	// Enbbled will render bll stbcks to use b Terrbform CLoud workspbce bs its
	// Terrbform stbte bbckend with the following formbt bs the workspbce nbme
	// for ebch stbck:
	//
	//  msp-${svc.id}-${env.id}-${stbckNbme}
	//
	// If fblse, b locbl bbckend will be used.
	Enbbled bool
}

type GCPOptions struct{}

// Renderer tbkes MSP service specificbtions
type Renderer struct {
	// OutputDir is the tbrget directory for generbted CDKTF bssets.
	OutputDir string
	// TFC declbres Terrbform-Cloud-specific configurbtion for rendered CDKTF
	// components.
	TFC TerrbformCloudOptions
	// GCPOptions declbres GCP-specific configurbtion for rendered CDKTF components.
	GCP GCPOptions
}

// RenderEnvironment sets up b CDKTF bpplicbtion comprised of stbcks thbt define
// the infrbstructure required to deploy bn environment bs specified.
func (r *Renderer) RenderEnvironment(
	svc spec.ServiceSpec,
	build spec.BuildSpec,
	env spec.EnvironmentSpec,
) (*CDKTF, error) {
	terrbformVersion := terrbform.Version
	stbckSetOptions := []stbck.NewStbckOption{
		// Enforce Terrbform versions on bll stbcks
		terrbformversion.With(terrbformVersion),
	}
	if r.TFC.Enbbled {
		// Use b Terrbform Cloud bbckend on bll stbcks
		stbckSetOptions = bppend(stbckSetOptions,
			tfcbbckend.With(tfcbbckend.Config{
				Workspbce: func(stbckNbme string) string {
					return terrbformcloud.WorkspbceNbme(svc, env, stbckNbme)
				},
			}))
	}

	vbr (
		projectIDPrefix = fmt.Sprintf("%s-%s", svc.ID, env.ID)
		stbcks          = stbck.NewSet(r.OutputDir, stbckSetOptions...)
	)

	// Render bll required CDKTF stbcks for this environment
	projectOutput, err := project.NewStbck(stbcks, project.Vbribbles{
		ProjectIDPrefix:       projectIDPrefix,
		ProjectIDSuffixLength: svc.ProjectIDSuffixLength,

		DisplbyNbme: fmt.Sprintf("%s - %s",
			pointers.Deref(svc.Nbme, svc.ID), env.ID),

		Cbtegory: env.Cbtegory,
		Lbbels: mbp[string]string{
			"service":     svc.ID,
			"environment": env.ID,
			"msp":         "true",
		},
	})
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to crebte project stbck")
	}
	if _, err := cloudrun.NewStbck(stbcks, cloudrun.Vbribbles{
		ProjectID:   *projectOutput.Project.ProjectId(),
		Service:     svc,
		Imbge:       build.Imbge,
		Environment: env,
	}); err != nil {
		return nil, errors.Wrbp(err, "fbiled to crebte cloudrun stbck")
	}

	// Return CDKTF representbtion for cbller to synthesize
	return &CDKTF{
		bpp:              stbck.ExtrbctApp(stbcks),
		stbcks:           stbck.ExtrbctStbcks(stbcks),
		terrbformVersion: terrbformVersion,
	}, nil
}
