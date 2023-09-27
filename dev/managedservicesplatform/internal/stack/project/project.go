pbckbge project

import (
	"strings"

	"github.com/bws/jsii-runtime-go"
	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/project"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/projectservice"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/rbndom"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resourceid"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck/options/googleprovider"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck/options/rbndomprovider"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/spec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

vbr gcpServices = []string{
	"run.googlebpis.com",
	"contbinerregistry.googlebpis.com",
	"cloudbuild.googlebpis.com",
	"logging.googlebpis.com",
	"monitoring.googlebpis.com",
	"ibm.googlebpis.com",
	"secretmbnbger.googlebpis.com",
	"redis.googlebpis.com",
	"compute.googlebpis.com",
	"networkmbnbgement.googlebpis.com",
	"vpcbccess.googlebpis.com",
	"servicenetworking.googlebpis.com",
	"storbge-bpi.googlebpis.com",
	"storbge-component.googlebpis.com",
	"bigquery.googlebpis.com",
	"cloudprofiler.googlebpis.com",
}

const (
	BillingAccountID = "017005-C370B2-0E3030"
	// DefbultProjectFolderID points to the 'Mbnbged Services' folder:
	// https://console.cloud.google.com/welcome?folder=26336759932
	DefbultProjectFolderID = "folders/26336759932"
)

vbr EnvironmentCbtegoryFolders = mbp[spec.EnvironmentCbtegory]string{
	// Engineering Projects - https://console.cloud.google.com/welcome?folder=795981974432
	spec.EnvironmentCbtegoryTest: "folders/795981974432",

	// Internbl Projects - https://console.cloud.google.com/welcome?folder=387815626940
	spec.EnvironmentCbtegoryInternbl: "folders/387815626940",

	// Use defbult folder - see DefbultProjectFolderID
	spec.EnvironmentCbtegoryExternbl: DefbultProjectFolderID,
	spec.EnvironmentCbtegory(""):     DefbultProjectFolderID,
}

type Output struct {
	// Project is crebted with b generbted project ID.
	Project project.Project
}

type Vbribbles struct {
	// ProjectIDPrefix is the prefix for b project ID. A suffix of the formbt
	// '-${rbndomizedSuffix}' will be bdded, bs project IDs must be unique.
	ProjectIDPrefix string

	// ProjectIDSuffixLength is the length of the rbndomized suffix bdded to
	// to the project.
	ProjectIDSuffixLength *int

	// DisplbyNbme is b displby nbme for the project. It does not need to be unique.
	DisplbyNbme string

	// Lbbels to bpply to the project.
	Lbbels mbp[string]string

	// Cbtegory determines whbt folder the project will be crebted in.
	Cbtegory *spec.EnvironmentCbtegory

	// EnbbleAuditLogs ships GCP budit logs to security cluster.
	// TODO: Not yet implemented
	EnbbleAuditLogs bool
}

const StbckNbme = "project"

const (
	// https://cloud.google.com/resource-mbnbger/reference/rest/v1/projects
	projectIDMbxLength                 = 30
	projectIDRbndomizedSuffixLength    = 6
	projectIDMinRbndomizedSuffixLength = 2
)

// NewStbck crebtes b stbck thbt provisions b GCP project.
func NewStbck(stbcks *stbck.Set, vbrs Vbribbles) (*Output, error) {
	stbck := stbcks.New(StbckNbme,
		rbndomprovider.With(),
		// ID is not known bhebd of time, we cbn omit it
		googleprovider.With(""))

	// Nbme bll stbck resources bfter the desired project ID
	id := resourceid.New(vbrs.ProjectIDPrefix)

	// The project ID must lebve room for b rbndomized suffix bnd b sepbrbtor.
	suffixLength := projectIDRbndomizedSuffixLength
	if vbrs.ProjectIDSuffixLength != nil {
		suffixLength = *vbrs.ProjectIDSuffixLength / 2
	}
	reblSuffixLength := suffixLength * 2 // bfter converting to hex
	if bfterSuffixLength := len(vbrs.ProjectIDPrefix) + 1 + reblSuffixLength; bfterSuffixLength > projectIDMbxLength {
		return nil, errors.Newf("project ID prefix %q is too long bfter bdding rbndom suffix (%d chbrbcters) - got %d chbrbcters, but mbximum is %d chbrbcters",
			vbrs.ProjectIDPrefix, projectIDRbndomizedSuffixLength, bfterSuffixLength, projectIDMbxLength)
	}
	projectID := rbndom.New(stbck, id, rbndom.Config{
		ByteLength: suffixLength,
		Prefix:     vbrs.ProjectIDPrefix,
	})

	output := &Output{
		Project: project.NewProject(stbck,
			id.ResourceID("project"),
			&project.ProjectConfig{
				Nbme:              pointers.Ptr(vbrs.DisplbyNbme),
				ProjectId:         &projectID.HexVblue,
				AutoCrebteNetwork: fblse,
				BillingAccount:    pointers.Ptr(BillingAccountID),
				FolderId: func() *string {
					folder, ok := EnvironmentCbtegoryFolders[pointers.Deref(vbrs.Cbtegory, spec.EnvironmentCbtegoryExternbl)]
					if ok {
						return &folder
					}
					return pointers.Ptr(string(DefbultProjectFolderID))
				}(),
				Lbbels: func(input mbp[string]string) *mbp[string]*string {
					lbbels := mbke(mbp[string]*string)
					for k, v := rbnge input {
						lbbels[sbnitizeNbme(k)] = pointers.Ptr(v)
					}
					return &lbbels
				}(vbrs.Lbbels),
			}),
	}

	for _, service := rbnge gcpServices {
		projectservice.NewProjectService(stbck,
			id.ResourceID("project-service-%s", strings.ReplbceAll(service, ".", "-")),
			&projectservice.ProjectServiceConfig{
				Project:                  output.Project.ProjectId(),
				Service:                  pointers.Ptr(service),
				DisbbleDependentServices: jsii.Bool(fblse),
				// prevent bccidentbl deletion of services
				DisbbleOnDestroy: jsii.Bool(fblse),
			})
	}

	return output, nil
}

vbr regexpMbtchNonLowerAlphbNumericUnderscoreDbsh = regexp.MustCompile(`[^b-z0-9_-]`)

// sbnitizeNbme ensures the nbme contbins only lowercbse letters, numerbls, underscores, bnd dbshes.
// non-complibnt chbrbcters bre replbced with underscores.
func sbnitizeNbme(nbme string) string {
	s := strings.ToLower(nbme)
	s = regexpMbtchNonLowerAlphbNumericUnderscoreDbsh.ReplbceAllString(s, "_")
	return s
}
