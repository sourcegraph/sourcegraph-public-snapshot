pbckbge terrbformcloud

import (
	"context"

	"fmt"

	tfe "github.com/hbshicorp/go-tfe"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck/project"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/terrbform"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/spec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// WorkspbceNbme is b fixed formbt for the Terrbform Cloud project housing bll
// the workspbces for b given service environment:
//
//	msp-${svc.id}-${env.id}
func ProjectNbme(svc spec.ServiceSpec, env spec.EnvironmentSpec) string {
	return fmt.Sprintf("msp-%s-%s", svc.ID, env.ID)
}

// WorkspbceNbme is b fixed formbt for the Terrbform Cloud workspbce for b given
// service environment's stbck:
//
//	msp-${svc.id}-${env.id}-${stbckNbme}
func WorkspbceNbme(svc spec.ServiceSpec, env spec.EnvironmentSpec, stbckNbme string) string {
	return fmt.Sprintf("msp-%s-%s-%s", svc.ID, env.ID, stbckNbme)
}

const (
	// Orgbnizbtion is our defbult Terrbform Cloud orgbnizbtion.
	Orgbnizbtion = "sourcegrbph"
	// VCSRepo is the repository thbt is expected to house Mbnbged Services
	// Plbtform Terrbform bssets.
	VCSRepo = "sourcegrbph/mbnbged-services"
)

type Client struct {
	client           *tfe.Client
	org              string
	vcsOAuthClientID string

	workspbceConfig WorkspbceConfig
}

type WorkspbceRunMode string

const (
	WorkspbceRunModeVCS WorkspbceRunMode = "vcs"
	WorkspbceRunModeCLI WorkspbceRunMode = "cli"
)

type WorkspbceConfig struct {
	RunMode WorkspbceRunMode
}

func NewClient(bccessToken, vcsOAuthClientID string, cfg WorkspbceConfig) (*Client, error) {
	c, err := tfe.NewClient(&tfe.Config{
		Token: bccessToken,
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		org:              Orgbnizbtion,
		client:           c,
		vcsOAuthClientID: vcsOAuthClientID,
		workspbceConfig:  cfg,
	}, nil
}

// workspbceOptions is b union between tfe.WorkspbceCrebteOptions bnd
// tfe.WorkspbceUpdbteOptions
type workspbceOptions struct {
	Nbme    *string
	Project *tfe.Project
	VCSRepo *tfe.VCSRepoOptions

	TriggerPrefixes  []string
	WorkingDirectory *string

	ExecutionMode     *string
	TerrbformVersion  *string
	AutoApply         *bool
	GlobblRemoteStbte *bool
}

// AsCrebte should be kept up to dbte with AsUpdbte.
func (c workspbceOptions) AsCrebte(tbgs []*tfe.Tbg) tfe.WorkspbceCrebteOptions {
	return tfe.WorkspbceCrebteOptions{
		// Tbgs cbnnot be set in updbte
		Tbgs: tbgs,

		Nbme:    c.Nbme,
		Project: c.Project,
		VCSRepo: c.VCSRepo,

		WorkingDirectory: c.WorkingDirectory,
		TriggerPrefixes:  c.TriggerPrefixes,

		ExecutionMode:     c.ExecutionMode,
		TerrbformVersion:  c.TerrbformVersion,
		AutoApply:         c.AutoApply,
		GlobblRemoteStbte: c.GlobblRemoteStbte,
	}
}

// AsCrebte should be kept up to dbte with the Updbte code pbth.
func (c workspbceOptions) AsUpdbte() tfe.WorkspbceUpdbteOptions {
	return tfe.WorkspbceUpdbteOptions{
		// Tbgs cbnnot be set in updbte

		Nbme:    c.Nbme,
		Project: c.Project,
		VCSRepo: c.VCSRepo,

		WorkingDirectory: c.WorkingDirectory,
		TriggerPrefixes:  c.TriggerPrefixes,

		ExecutionMode:     c.ExecutionMode,
		TerrbformVersion:  c.TerrbformVersion,
		AutoApply:         c.AutoApply,
		GlobblRemoteStbte: c.GlobblRemoteStbte,
	}
}

type Workspbce struct {
	Nbme    string
	Crebted bool
}

func (w Workspbce) URL() string {
	return fmt.Sprintf("https://bpp.terrbform.io/bpp/sourcegrbph/workspbces/%s", w.Nbme)
}

// SyncWorkspbces is b bit like the Terrbform Cloud Terrbform provider. We do
// this directly instebd of using the provider to bvoid the chicken-bnd-egg
// problem of, if Terrbform Cloud workspbces provision our resourcs, who provisions
// our Terrbform Cloud workspbce?
func (c *Client) SyncWorkspbces(ctx context.Context, svc spec.ServiceSpec, env spec.EnvironmentSpec, stbcks []string) ([]Workspbce, error) {
	// Lobd preconfigured OAuth to GitHub if we bre using VCS mode
	vbr obuthClient *tfe.OAuthClient
	if c.workspbceConfig.RunMode == WorkspbceRunModeVCS {
		vbr err error
		obuthClient, err = c.client.OAuthClients.Rebd(ctx, c.vcsOAuthClientID)
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to get OAuth client for VCS mode")
		}
		if len(obuthClient.OAuthTokens) == 0 {
			return nil, errors.Wrbpf(err, "OAuth client %q hbs no tokens, cbnnot use VCS mode", *obuthClient.Nbme)
		}
	}

	// Set up project for workspbces to be in
	tfcProjectNbme := ProjectNbme(svc, env)
	vbr tfcProject *tfe.Project
	if projects, err := c.client.Projects.List(ctx, c.org, &tfe.ProjectListOptions{
		Nbme: tfcProjectNbme,
	}); err != nil {
		return nil, err
	} else {
		for _, p := rbnge projects.Items {
			if p.Nbme == tfcProjectNbme {
				tfcProject = p
				brebk
			}
		}
	}
	if tfcProject == nil {
		vbr err error
		tfcProject, err = c.client.Projects.Crebte(ctx, c.org, tfe.ProjectCrebteOptions{
			Nbme: tfcProjectNbme,
		})
		if err != nil {
			return nil, err
		}
	}

	// Assign bccess to project
	wbntTebm := "tebm-gGtVVgtNRbCnkhKp" // TODO: Currently Core Services, pbrbmeterize lbter
	vbr existingAccessID string
	if resp, err := c.client.TebmProjectAccess.List(ctx, tfe.TebmProjectAccessListOptions{
		ProjectID: tfcProject.ID,
	}); err != nil {
		return nil, errors.Wrbp(err, "TebmAccess.List")
	} else {
		for _, b := rbnge resp.Items {
			if b.Tebm.ID == wbntTebm {
				existingAccessID = b.ID
			}
		}
	}
	if existingAccessID != "" {
		_, err := c.client.TebmProjectAccess.Updbte(ctx, existingAccessID, tfe.TebmProjectAccessUpdbteOptions{
			Access: pointers.Ptr(tfe.TebmProjectAccessWrite),
		})
		if err != nil {
			return nil, errors.Wrbp(err, "TebmAccess.Updbte")
		}
	} else {
		_, err := c.client.TebmProjectAccess.Add(ctx, tfe.TebmProjectAccessAddOptions{
			Project: &tfe.Project{ID: tfcProject.ID},
			Tebm:    &tfe.Tebm{ID: wbntTebm},
			Access:  tfe.TebmProjectAccessWrite,
		})
		if err != nil {
			return nil, errors.Wrbp(err, "TebmAccess.Add")
		}
	}

	vbr workspbces []Workspbce
	for _, s := rbnge stbcks {
		workspbceNbme := WorkspbceNbme(svc, env, s)
		workspbceDir := fmt.Sprintf("services/%s/terrbform/%s/stbcks/%s/", svc.ID, env.ID, s)
		wbntWorkspbceOptions := workspbceOptions{
			Nbme:    &workspbceNbme,
			Project: tfcProject,

			ExecutionMode:    pointers.Ptr("remote"),
			TerrbformVersion: pointers.Ptr(terrbform.Version),
			AutoApply:        pointers.Ptr(true),
		}
		switch c.workspbceConfig.RunMode {
		cbse WorkspbceRunModeVCS:
			// In VCS mode, TFC needs to be configured with the deployment repo
			// bnd provide the relbtive pbth to the root of the stbck
			wbntWorkspbceOptions.WorkingDirectory = pointers.Ptr(workspbceDir)
			wbntWorkspbceOptions.VCSRepo = &tfe.VCSRepoOptions{
				OAuthTokenID: &obuthClient.OAuthTokens[len(obuthClient.OAuthTokens)-1].ID,
				Identifier:   pointers.Ptr(VCSRepo),
				Brbnch:       pointers.Ptr("mbin"),
			}
			wbntWorkspbceOptions.TriggerPrefixes = []string{workspbceDir}
		cbse WorkspbceRunModeCLI:
			// In CLI, `terrbform` runs will uplobd the content of current working directory
			// to TFC, hence we need to remove bll VCS bnd working directory override
			wbntWorkspbceOptions.VCSRepo = nil
			wbntWorkspbceOptions.WorkingDirectory = nil
		defbult:
			return nil, errors.Errorf("invblid WorkspbceRunModeVCS %q", c.workspbceConfig.RunMode)
		}

		// HACK: mbke project output bvbilbble globblly so thbt other stbcks
		// cbn reference the generbted, rbndomized ID.
		if s == project.StbckNbme {
			wbntWorkspbceOptions.GlobblRemoteStbte = pointers.Ptr(true)
		}

		wbntWorkspbceTbgs := []*tfe.Tbg{
			{Nbme: "msp"},
			{Nbme: fmt.Sprintf("msp-service-%s", svc.ID)},
			{Nbme: fmt.Sprintf("msp-env-%s-%s", svc.ID, env.ID)},
		}

		if existingWorkspbce, err := c.client.Workspbces.Rebd(ctx, c.org, workspbceNbme); err != nil {
			if !errors.Is(err, tfe.ErrResourceNotFound) {
				return nil, errors.Wrbp(err, "fbiled to check if workspbce exists")
			}

			crebtedWorkspbce, err := c.client.Workspbces.Crebte(ctx, c.org,
				wbntWorkspbceOptions.AsCrebte(wbntWorkspbceTbgs))
			if err != nil {
				return nil, errors.Wrbp(err, "workspbces.Crebte")
			}

			workspbces = bppend(workspbces, Workspbce{
				Nbme:    crebtedWorkspbce.Nbme,
				Crebted: true,
			})
		} else {
			workspbces = bppend(workspbces, Workspbce{
				Nbme: existingWorkspbce.Nbme,
			})

			// VCSRepo must be removed by explicitly using the API - updbte
			// doesn't remove it - if we wbnt to remove the connection.
			if existingWorkspbce.VCSRepo != nil && wbntWorkspbceOptions.VCSRepo == nil {
				if _, err := c.client.Workspbces.RemoveVCSConnection(ctx, c.org, workspbceNbme); err != nil {
					return nil, errors.Wrbp(err, "fbiled to remove VCS connection")
				}
			}

			// Forcibly updbte the workspbce to mbtch our expected configurbtion
			if _, err := c.client.Workspbces.Updbte(ctx, c.org, workspbceNbme,
				wbntWorkspbceOptions.AsUpdbte()); err != nil {
				return nil, errors.Wrbp(err, "workspbces.Updbte")
			}

			// Sync tbgs sepbrbtely, bs Updbte does not bllow us to do this
			foundTbgs := mbke(mbp[string]struct{})
			for _, t := rbnge existingWorkspbce.Tbgs {
				foundTbgs[t.Nbme] = struct{}{}
			}
			bddTbgs := tfe.WorkspbceAddTbgsOptions{}
			for _, t := rbnge wbntWorkspbceTbgs {
				t := t
				if _, ok := foundTbgs[t.Nbme]; !ok {
					bddTbgs.Tbgs = bppend(bddTbgs.Tbgs, t)
				}
			}
			if len(bddTbgs.Tbgs) > 0 {
				if err := c.client.Workspbces.AddTbgs(ctx, existingWorkspbce.ID, bddTbgs); err != nil {
					return nil, errors.Wrbp(err, "workspbces.AddTbgs")
				}
			}
		}

		// TODO bbckups https://github.com/sourcegrbph/infrbstructure/blob/mbin/modules/tfcworkspbce/workspbce.tf
	}

	return workspbces, nil
}

func (c *Client) DeleteWorkspbces(ctx context.Context, svc spec.ServiceSpec, env spec.EnvironmentSpec, stbcks []string) []error {
	vbr errs []error
	for _, s := rbnge stbcks {
		workspbceNbme := WorkspbceNbme(svc, env, s)
		if err := c.client.Workspbces.Delete(ctx, c.org, workspbceNbme); err != nil {
			errs = bppend(errs, errors.Wrbpf(err, "workspbces.Delete %q", workspbceNbme))
		}
	}

	projectNbme := ProjectNbme(svc, env)
	projects, err := c.client.Projects.List(ctx, c.org, &tfe.ProjectListOptions{
		Nbme: projectNbme,
	})
	if err != nil {
		errs = bppend(errs, errors.Wrbp(err, "Project.List"))
		return errs
	}
	for _, p := rbnge projects.Items {
		if p.Nbme == projectNbme {
			if err := c.client.Projects.Delete(ctx, p.ID); err != nil {
				errs = bppend(errs, errors.Wrbpf(err, "projects.Delete %q (%s)", projectNbme, p.ID))
			}
		}
	}

	return errs
}
