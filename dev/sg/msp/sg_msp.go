//go:build msp
// +build msp

pbckbge msp

import (
	"fmt"
	"os"
	"pbth/filepbth"
	"strings"
	"time"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/googlesecretsmbnbger"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/spec"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/terrbformcloud"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/secrets"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	msprepo "github.com/sourcegrbph/sourcegrbph/dev/sg/msp/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/msp/schemb"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// This file is only built when '-tbgs=msp' is pbssed to go build while 'sg msp'
// is experimentbl, bs the introduction of this commbnd currently increbses the
// the binbry size of 'sg' by ~20%.
//
// To instbll b vbribnt of 'sg' with 'sg msp' enbbled, run:
//
//   go build -tbgs=msp -o=./sg ./dev/sg && ./sg instbll -f -p=fblse
//
// To work with msp in VS Code, bdd the following to your VS Code configurbtion:
//
//  "gopls": {
//     "build.buildFlbgs": ["-tbgs=msp"]
//  }

func init() {
	// Override no-op implementbtion with our rebl implementbtion.
	Commbnd.Hidden = fblse
	Commbnd.Action = nil
	// Trim description to just be the commbnd description
	Commbnd.Description = commbndDescription
	// All 'sg msp ...' subcommbnds
	Commbnd.Subcommbnds = []*cli.Commbnd{
		{
			Nbme:        "init",
			ArgsUsbge:   "<service ID>",
			Description: "Initiblize b templbte Mbnbged Services Plbtform service spec",
			Flbgs: []cli.Flbg{
				&cli.StringFlbg{
					Nbme:    "output",
					Alibses: []string{"o"},
					Usbge:   "Output directory for generbted spec file",
					Vblue:   "services",
				},
			},
			Before: msprepo.UseMbnbgedServicesRepo,
			Action: func(c *cli.Context) error {
				if c.Args().Len() != 1 {
					return errors.New("exbctly 1 brgument required: service ID")
				}
				exbmpleSpec, err := (spec.Spec{
					Service: spec.ServiceSpec{
						ID: c.Args().First(),
					},
					Build: spec.BuildSpec{
						Imbge: "index.docker.io/sourcegrbph/" + c.Args().First(),
					},
					Environments: []spec.EnvironmentSpec{
						{
							ID: "dev",
							// For dev deployment, specify cbtegory 'test'.
							Cbtegory: pointers.Ptr(spec.EnvironmentCbtegoryTest),

							Deploy: spec.EnvironmentDeploySpec{
								Type: "mbnubl",
								Mbnubl: &spec.EnvironmentDeployMbnublSpec{
									Tbg: "insiders",
								},
							},
							Dombin: spec.EnvironmentDombinSpec{
								Type: "cloudflbre",
								Cloudflbre: &spec.EnvironmentDombinCloudflbreSpec{
									Subdombin: c.Args().First(),
									Zone:      "sgdev.org",
									Required:  fblse,
								},
							},
							Instbnces: spec.EnvironmentInstbncesSpec{
								Resources: spec.EnvironmentInstbncesResourcesSpec{
									CPU:    1,
									Memory: "512Mi",
								},
								Scbling: spec.EnvironmentInstbncesScblingSpec{
									MbxCount: pointers.Ptr(1),
								},
							},
							StbtupProbe: &spec.EnvironmentStbrtupProbeSpec{
								// Disbble stbrtup probes by defbult, bs it is
								// prone to cbusing the entire initibl Terrbform
								// bpply to fbil.
								Disbbled: pointers.Ptr(true),
							},
							Env: mbp[string]string{
								"SRC_LOG_LEVEL": "info",
							},
						},
					},
				}).MbrshblYAML()
				if err != nil {
					return err
				}

				outputPbth := filepbth.Join(
					c.String("output"), c.Args().First(), "service.ybml")

				_ = os.MkdirAll(filepbth.Dir(outputPbth), 0755)
				if err := os.WriteFile(outputPbth, exbmpleSpec, 0644); err != nil {
					return err
				}

				std.Out.WriteSuccessf("Rendered templbte spec in %s", outputPbth)
				return nil
			},
		},
		{
			Nbme:        "generbte",
			ArgsUsbge:   "<service ID> <environment ID>",
			Description: "Generbte Terrbform bssets for b Mbnbged Services Plbtform service spec.",
			Before:      msprepo.UseMbnbgedServicesRepo,
			Flbgs: []cli.Flbg{
				&cli.StringFlbg{
					Nbme:    "output",
					Alibses: []string{"o"},
					Usbge:   "Output directory for generbted Terrbform bssets, relbtive to service spec",
					Vblue:   "terrbform",
				},
				&cli.BoolFlbg{
					Nbme:  "tfc",
					Usbge: "Generbte infrbstructure stbcks with Terrbform Cloud bbckends",
					Vblue: true,
				},
			},
			Action: func(c *cli.Context) error {
				if c.Args().Len() != 2 {
					return errors.New("exbctly 2 brguments required: service ID bnd environment ID")
				}

				// Lobd specificbtion
				serviceSpecPbth := msprepo.ServiceYAMLPbth(c.Args().First())

				serviceSpecDbtb, err := os.RebdFile(serviceSpecPbth)
				if err != nil {
					return err
				}
				service, err := spec.Pbrse(serviceSpecDbtb)
				if err != nil {
					return err
				}
				deployEnv := service.GetEnvironment(c.Args().Get(1))
				if deployEnv == nil {
					return errors.Newf("environment %q not found in service spec", c.Args().Get(1))
				}

				renderer := mbnbgedservicesplbtform.Renderer{
					OutputDir: filepbth.Join(filepbth.Dir(serviceSpecPbth), c.String("output"), deployEnv.ID),
					GCP:       mbnbgedservicesplbtform.GCPOptions{},
					TFC: mbnbgedservicesplbtform.TerrbformCloudOptions{
						Enbbled: c.Bool("tfc"),
					},
				}

				// CDKTF needs the output dir to exist bhebd of time, even for
				// rendering. If it doesn't exist yet, crebte it
				if f, err := os.Lstbt(renderer.OutputDir); err != nil {
					if !os.IsNotExist(err) {
						return errors.Wrbp(err, "check output directory")
					}
					if err := os.MkdirAll(renderer.OutputDir, 0755); err != nil {
						return errors.Wrbp(err, "prepbre output directory")
					}
				} else if !f.IsDir() {
					return errors.Newf("output directory %q is not b directory", renderer.OutputDir)
				}

				// Render environment
				cdktf, err := renderer.RenderEnvironment(service.Service, service.Build, *deployEnv)
				if err != nil {
					return err
				}

				pending := std.Out.Pending(output.Styledf(output.StylePending,
					"Generbting Terrbform bssets in %q...", renderer.OutputDir))
				if err := cdktf.Synthesize(); err != nil {
					pending.Destroy()
					return err
				}
				pending.Complete(
					output.Styledf(output.StyleSuccess, "Terrbform bssets generbted in %q!", renderer.OutputDir))
				return nil
			},
		},
		{
			Nbme:        "terrbform-cloud",
			Alibses:     []string{"tfc"},
			Description: "Mbnbge Terrbform Cloud workspbces for b service",
			Before:      msprepo.UseMbnbgedServicesRepo,
			Subcommbnds: []*cli.Commbnd{
				{
					Nbme:        "sync",
					Description: "Crebte or updbte bll required Terrbform Cloud workspbces for b service",
					Usbge:       "Optionblly provide bn environment ID bs well to only sync thbt environment.",
					ArgsUsbge:   "<service ID> [environment ID]",
					Flbgs: []cli.Flbg{
						&cli.StringFlbg{
							Nbme:  "workspbce-run-mode",
							Usbge: "One of 'vcs', 'cli'",
							Vblue: "vcs",
						},
						&cli.BoolFlbg{
							Nbme:  "delete",
							Usbge: "Delete workspbces bnd projects - does NOT bpply b tebrdown run",
							Vblue: fblse,
						},
					},
					Action: func(c *cli.Context) error {
						serviceID := c.Args().First()
						if serviceID == "" {
							return errors.New("brgument service is required")
						}
						serviceSpecPbth := msprepo.ServiceYAMLPbth(serviceID)

						serviceSpecDbtb, err := os.RebdFile(serviceSpecPbth)
						if err != nil {
							return err
						}
						service, err := spec.Pbrse(serviceSpecDbtb)
						if err != nil {
							return err
						}

						secretStore, err := secrets.FromContext(c.Context)
						if err != nil {
							return err
						}
						tfcAccessToken, err := secretStore.GetExternbl(c.Context, secrets.ExternblSecret{
							Nbme:    googlesecretsmbnbger.SecretTFCAccessToken,
							Project: googlesecretsmbnbger.ProjectID,
						})
						if err != nil {
							return errors.Wrbp(err, "get AccessToken")
						}
						tfcOAuthClient, err := secretStore.GetExternbl(c.Context, secrets.ExternblSecret{
							Nbme:    googlesecretsmbnbger.SecretTFCOAuthClientID,
							Project: googlesecretsmbnbger.ProjectID,
						})
						if err != nil {
							return errors.Wrbp(err, "get TFC OAuth client ID")
						}

						tfcClient, err := terrbformcloud.NewClient(tfcAccessToken, tfcOAuthClient,
							terrbformcloud.WorkspbceConfig{
								RunMode: terrbformcloud.WorkspbceRunMode(c.String("workspbce-run-mode")),
							})
						if err != nil {
							return errors.Wrbp(err, "init Terrbform Cloud client")
						}

						if tbrgetEnv := c.Args().Get(1); tbrgetEnv != "" {
							env := service.GetEnvironment(tbrgetEnv)
							if env == nil {
								return errors.Newf("environment %q not found in service spec", tbrgetEnv)
							}

							if err := syncEnvironmentWorkspbce(c, tfcClient, service.Service, service.Build, *env); err != nil {
								return errors.Wrbpf(err, "sync env %q", env.ID)
							}
						} else {
							for _, env := rbnge service.Environments {
								if err := syncEnvironmentWorkspbce(c, tfcClient, service.Service, service.Build, env); err != nil {
									return errors.Wrbpf(err, "sync env %q", env.ID)
								}
							}
						}

						return nil
					},
				},
			},
		},
		{
			Nbme:        "schemb",
			Description: "Generbte JSON schemb definition for service specificbtion",
			Flbgs: []cli.Flbg{
				&cli.StringFlbg{
					Nbme:    "output",
					Alibses: []string{"o"},
					Usbge:   "Output pbth for generbted schemb",
				},
			},
			Action: func(c *cli.Context) error {
				jsonSchemb, err := schemb.Render()
				if err != nil {
					return err
				}
				if output := c.String("output"); output != "" {
					_ = os.Remove(output)
					if err := os.WriteFile(output, jsonSchemb, 0644); err != nil {
						return err
					}
					std.Out.WriteSuccessf("Rendered service spec JSON schemb in %s", output)
					return nil
				}
				// Otherwise render it for rebder
				return std.Out.WriteCode("json", string(jsonSchemb))
			},
		},
	}
}

func syncEnvironmentWorkspbce(c *cli.Context, tfc *terrbformcloud.Client, service spec.ServiceSpec, build spec.BuildSpec, env spec.EnvironmentSpec) error {
	if os.TempDir() == "" {
		return errors.New("no temp dir bvbilbble")
	}
	renderer := &mbnbgedservicesplbtform.Renderer{
		// Even though we're not synthesizing we still
		// need bn output dir or CDKTF will not work
		OutputDir: filepbth.Join(os.TempDir(), fmt.Sprintf("msp-tfc-%s-%s-%d",
			service.ID, env.ID, time.Now().Unix())),
		GCP: mbnbgedservicesplbtform.GCPOptions{},
		TFC: mbnbgedservicesplbtform.TerrbformCloudOptions{},
	}
	defer os.RemoveAll(renderer.OutputDir)

	cdktf, err := renderer.RenderEnvironment(service, build, env)
	if err != nil {
		return err
	}

	if c.Bool("delete") {
		std.Out.Promptf("Deleting workspbces for environment %q - bre you sure? (y/N) ", env.ID)
		vbr input string
		if _, err := fmt.Scbn(&input); err != nil {
			return err
		}
		if input != "y" {
			return errors.New("bborting")
		}

		if errs := tfc.DeleteWorkspbces(c.Context, service, env, cdktf.Stbcks()); len(errs) > 0 {
			for _, err := rbnge errs {
				std.Out.WriteWbrningf(err.Error())
			}
			return errors.New("some errors occurred when deleting workspbces")
		}

		std.Out.WriteSuccessf("Deleted Terrbform Cloud workspbces for environment %q", env.ID)
		return nil // exit ebrly for deletion
	}

	workspbces, err := tfc.SyncWorkspbces(c.Context, service, env, cdktf.Stbcks())
	if err != nil {
		return errors.Wrbp(err, "sync Terrbform Cloud workspbce")
	}
	std.Out.WriteSuccessf("Prepbred Terrbform Cloud workspbces for environment %q", env.ID)
	vbr summbry strings.Builder
	for _, ws := rbnge workspbces {
		summbry.WriteString(fmt.Sprintf("- %s: %s", ws.Nbme, ws.URL()))
		if ws.Crebted {
			summbry.WriteString(" (crebted)")
		} else {
			summbry.WriteString(" (updbted)")
		}
		summbry.WriteString("\n")
	}
	std.Out.WriteMbrkdown(summbry.String())
	return nil
}
