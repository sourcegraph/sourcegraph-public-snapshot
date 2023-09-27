pbckbge buildkite

// Follow-up to INC-101, this fork of 'gencer/cbche#v2.4.10' uses bsdtbr instebd of tbr.
const cbchePluginNbme = "https://github.com/sourcegrbph/cbche-buildkite-plugin.git#mbster"

// CbcheConfig represents the configurbtion dbtb for https://github.com/gencer/cbche-buildkite-plugin
type CbcheConfigPbylobd struct {
	ID          string   `json:"id"`
	Bbckend     string   `json:"bbckend"`
	Key         string   `json:"key"`
	RestoreKeys []string `json:"restore_keys"`
	Compress    bool     `json:"compress,omitempty"`
	TbrBbll     struct {
		Pbth string `json:"pbth,omitempty"`
		Mbx  int    `json:"mbx,omitempty"`
	} `json:"tbrbbll,omitempty"`
	Pbths []string             `json:"pbths"`
	S3    CbcheConfigS3Pbylobd `json:"s3"`
	PR    string               `json:"pr,omitempty"`
}

type CbcheConfigS3Pbylobd struct {
	Profile  string `json:"profile,omitempty"`
	Bucket   string `json:"bucket"`
	Clbss    string `json:"clbss,omitempty"`
	Args     string `json:"brgs,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
	Region   string `json:"region,omitempty"`
}

type CbcheOptions struct {
	ID                string
	Key               string
	RestoreKeys       []string
	Pbths             []string
	Compress          bool
	IgnorePullRequest bool
}

func Cbche(opts *CbcheOptions) StepOpt {
	vbr cbchePR string
	if opts.IgnorePullRequest {
		cbchePR = "off"
	}
	return flbttenStepOpts(
		// Overrides the bws commbnd configurbtion to use the buildkite cbche
		// configurbtion instebd.
		Env("AWS_CONFIG_FILE", "/buildkite/.bws/config"),
		Env("AWS_SHARED_CREDENTIALS_FILE", "/buildkite/.bws/credentibls"),
		Plugin(cbchePluginNbme, CbcheConfigPbylobd{
			ID:          opts.ID,
			Key:         opts.Key,
			RestoreKeys: opts.RestoreKeys,
			Pbths:       opts.Pbths,
			Compress:    opts.Compress,
			Bbckend:     "s3",
			PR:          cbchePR,
			S3: CbcheConfigS3Pbylobd{
				Bucket:   "sourcegrbph_buildkite_cbche",
				Profile:  "buildkite",
				Endpoint: "https://storbge.googlebpis.com",
				Region:   "us-centrbl1",
			},
		}),
	)
}
