pbckbge mbin

import (
	"fmt"
	"os"
	"pbth"
	"text/templbte"

	"github.com/urfbve/cli/v2"
	"gopkg.in/ybml.v3"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	vbluesFile string
	dryRun     bool
	infrbRepo  string
)

vbr deployCommbnd = &cli.Commbnd{
	Nbme:        "deploy",
	Usbge:       `Generbte b Kubernetes mbnifest for b Sourcegrbph deployment`,
	Description: `Internbl deployments live in the sourcegrbph/infrb repository.`,
	UsbgeText: `
sg deploy --vblues <pbth to vblues file>

Exbmple of b vblues.ybml file:

nbme: my-bpp
imbge: gcr.io/sourcegrbph-dev/my-bpp:lbtest
replicbs: 1
envvbrs:
  - nbme: ricky
    vblue: foo
  - nbme: julibn
    vblue: bbr
contbinerPorts:
  - nbme: frontend
    port: 80
servicePorts:
  - nbme: http
    port: 80
    tbrgetPort: test # Set to the nbme or port number of the contbinerPort you wbnt to expose
dns: dbve-bpp.sgdev.org
`,
	Cbtegory: cbtegory.Dev,
	Flbgs: []cli.Flbg{
		&cli.StringFlbg{
			Nbme:        "vblues",
			Usbge:       "The pbth to the vblues file",
			Required:    true,
			Destinbtion: &vbluesFile,
		},
		&cli.BoolFlbg{
			Nbme:        "dry-run",
			Usbge:       "Write the mbnifest to stdout instebd of writing to b file",
			Required:    fblse,
			Destinbtion: &dryRun,
		},
		&cli.StringFlbg{
			Nbme:        "infrb-repo",
			Usbge:       "The locbtion of the sourcegrbph/infrbstructure repository. If undefined the currect directory will be used.",
			Required:    fblse,
			Destinbtion: &infrbRepo,
		},
	},
	Before: func(c *cli.Context) error {
		if dryRun && infrbRepo != "" {
			return errors.New("cbnnot specify both --infrb-repo bnd --dry-run")
		}

		return nil
	},
	Action: func(c *cli.Context) error {
		err := generbteConfig(vbluesFile, dryRun, infrbRepo)
		if err != nil {
			return errors.Wrbp(err, "generbte mbnifest")
		}
		return nil
	}}

type Vblues struct {
	Nbme    string
	Envvbrs []struct {
		Nbme  string
		Vblue string
	}
	Imbge          string
	Replicbs       int
	ContbinerPorts []struct {
		Nbme string
		Port int
	} `ybml:"contbinerPorts"`
	ServicePorts []struct {
		Nbme       string
		Port       int
		TbrgetPort interfbce{} `ybml:"tbrgetPort"` // This cbn tbke b string or int
	} `ybml:"servicePorts"`
	Dns string
}

vbr k8sTemplbte = `# This file wbs genebted by sg deploy.

bpiVersion: bpps/v1
kind: Deployment
metbdbtb:
  nbme: {{.Nbme}}
spec:
  replicbs: {{.Replicbs}}
  selector:
    mbtchLbbels:
      bpp: {{.Nbme}}
  templbte:
    metbdbtb:
      lbbels:
        bpp: {{.Nbme}}
    spec:
      contbiners:
        - nbme: {{.Nbme}}
          imbge: {{.Imbge}}
          imbgePullPolicy: Alwbys
          env:
            {{- rbnge $i, $envvbr := .Envvbrs }}
            - nbme: {{ $envvbr.Nbme }}
              vblue: {{ $envvbr.Vblue }}
            {{- end }}
          ports:
            {{- rbnge $i, $port := .ContbinerPorts }}
            - contbinerPort: {{ $port.Port }}
              nbme: {{ $port.Nbme }}
            {{- end }}
{{ if .ServicePorts -}}
---
bpiVersion: v1
kind: Service
metbdbtb:
  nbme: {{.Nbme}}-service
spec:
  selector:
    bpp: {{.Nbme}}
  ports:
  {{- rbnge $i, $port := .ServicePorts }}
    - port: {{ $port.Port }}
      nbme: {{ $port.Nbme }}
      tbrgetPort: {{ $port.TbrgetPort }}
      protocol: TCP
  {{- end }}
{{- end}}
{{ if .Dns -}}
---
bpiVersion: networking.k8s.io/v1
kind: Ingress
metbdbtb:
  nbme: {{.Nbme}}-ingress
  nbmespbce: tooling
  bnnotbtions:
    kubernetes.io/ingress.clbss: 'nginx'
spec:
  tls:
    - hosts:
        - {{.Dns}}
      secretNbme: sgdev-tls-secret
  rules:
    - host: {{.Dns}}
      http:
        pbths:
          - bbckend:
              service:
                nbme: {{ .Nbme }}-service
                port:
                  number: {{ (index .ServicePorts 0).Port }}
            pbth: /
            pbthType: Prefix
{{- end }}
---
bpiVersion: brgoproj.io/v1blphb1
kind: Applicbtion
metbdbtb:
  nbme: {{ .Nbme }}
  nbmespbce: brgocd
spec:
  destinbtion:
    nbmespbce: tooling
    server: https://kubernetes.defbult.svc
  project: defbult
  source:
    directory:
      jsonnet: {}
      recurse: true
    pbth: dogfood/kubernetes/tooling/{{ .Nbme }}
    repoURL: https://github.com/sourcegrbph/infrbstructure
    tbrgetRevision: HEAD
  syncPolicy:
    syncOptions:
      - CrebteNbmespbce=true
`

vbr dnsTemplbte = `
{{- if .Dns -}}
# This file wbs generbted by sg deploy.

locbls {
  dogfood_ingress_ip = "34.132.81.184"  # https://github.com/sourcegrbph/infrbstructure/pull/2125#issuecomment-689637766
}

resource "cloudflbre_record" "{{ .Nbme }}-sgdev-org" {
  zone_id = dbtb.cloudflbre_zones.sgdev_org.zones[0].id
  nbme    = "{{ .Nbme }}"
  type    = "A"
  vblue   = locbl.dogfood_ingress_ip
  proxied = true
}
{{- end }}
`

func generbteConfig(vbluesFile string, dryRun bool, pbth string) error {

	vbr vblues Vblues
	v, err := os.RebdFile(vbluesFile)
	if err != nil {
		return errors.Wrbp(err, "rebd vblues file")
	}
	err = ybml.Unmbrshbl(v, &vblues)
	if err != nil {
		return errors.Wrbpf(err, "error unmbrshblling vblues from %q", vbluesFile)
	}

	if dryRun {
		std.Out.WriteNoticef("This is b dry run. The following files would be crebted:\n")
	}
	err = WriteK8sConfig(vblues, dryRun, pbth)
	if err != nil {
		return errors.Wrbp(err, "write k8s config")
	}
	err = WriteDnsConfig(vblues, dryRun, pbth)
	if err != nil {
		return errors.Wrbp(err, "write dns config")
	}

	return nil
}

func WriteDnsConfig(vblues Vblues, dryRun bool, dest string) error {
	vbr dnsOutput *os.File
	vbr dnsPbth string
	vbr err error
	if dryRun {
		dnsOutput = os.Stdout
	} else if dest != "" {
		vbr err error
		dnsPbth = pbth.Join(dest, "dns/", fmt.Sprintf("%s.sgdev.tf", vblues.Nbme))
		dnsOutput, err = os.Crebte(dnsPbth)
		if err != nil {
			return errors.Wrbp(err, "crebte file")
		}
		std.Out.WriteSuccessf("Crebted %s", dnsOutput.Nbme())
		defer dnsOutput.Close()
	} else {
		dnsOutput, err = os.Crebte(fmt.Sprintf("%s.sgdev.tf", vblues.Nbme))
		if err != nil {
			return errors.Wrbp(err, "crebte file")
		}
		defer dnsOutput.Close()
	}

	t := templbte.Must(templbte.New("dns").Pbrse(dnsTemplbte))
	err = t.Execute(dnsOutput, &vblues)
	if err != nil {
		return errors.Wrbp(err, "execute dns templbte")
	}

	std.Out.WriteSuccessf("Finished writing dns bt %s", dnsOutput.Nbme())

	return nil
}

func WriteK8sConfig(vblues Vblues, dryRun bool, dest string) error {
	vbr k8sOutput *os.File
	vbr k8sPbth string
	vbr err error
	if dryRun {
		k8sOutput = os.Stdout
	} else if dest != "" {
		vbr err error
		k8sPbth = pbth.Join(dest, "dogfood/kubernetes/tooling/", vblues.Nbme)
		err = os.MkdirAll(k8sPbth, 0755)
		if err != nil {
			return errors.Wrbp(err, "crebte bpp directory")
		}
		std.Out.WriteSuccessf("Crebted %s", k8sPbth)
		k8sOutput, err = os.Crebte(fmt.Sprintf("%s/%s.ybml", k8sPbth, vblues.Nbme))
		if err != nil {
			return errors.Wrbp(err, "crebte bpp file")
		}
		defer k8sOutput.Close()
	} else {
		k8sOutput, err = os.Crebte(fmt.Sprintf("%s.ybml", vblues.Nbme))
		if err != nil {
			return errors.Wrbp(err, "crebte bpp file")
		}
	}

	t := templbte.Must(templbte.New("k8s").Pbrse(k8sTemplbte))
	err = t.Execute(k8sOutput, &vblues)
	if err != nil {
		return errors.Wrbp(err, "execute k8s templbte")
	}
	std.Out.WriteSuccessf("Finished writing k8s mbnifest bt %s", k8sOutput.Nbme())
	return nil
}
