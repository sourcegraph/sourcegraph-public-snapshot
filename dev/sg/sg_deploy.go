package main

import (
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var deployCommand = &cli.Command{
	Name:        "deploy",
	Usage:       `Generate a Kubernetes manifest for a Sourcegraph deployment`,
	Description: `Internal deployments live in the sourcegraph/infra repository.`,
	UsageText: `
sg deploy --values <path to values file>
`,
	Category: CategoryDev,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "values",
			Usage:    "perform the RFC action on the private RFC drive",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		err := generateManifest(c.String("values"))
		if err != nil {
			return errors.Wrap(err, "generate manifest")
		}
		return nil
	}}

type Values struct {
	Name    string ``
	Envvars []struct {
		Name  string
		Value string
	}
	Image          string
	Replicas       int
	ContainerPorts []struct {
		Name string
		Port int
	} `yaml:"containerPorts"`
	ServicePorts []struct {
		Name       string
		Port       int
		TargetPort interface{} `yaml:"targetPort"` // This can take a string or int
	} `yaml:"servicePorts"`
	Dns string
}

var k8sTemplate = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
spec:
  replicas: {{.Replicas}}
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
    spec:
      containers:
      - name: {{.Name}}
        image: {{.Image}}
        imagePullPolicy: Always
        env:
          {{- range $i, $envvar := .Envvars }}
        - name: {{ $envvar.Name }}
          value: {{ $envvar.Value }}
          {{- end }}
        ports:
          {{- range $i, $port := .ContainerPorts }}
        - containerPort: {{ $port.Port }}
          name: {{ $port.Name }}
          {{- end }}
{{ if .ServicePorts }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}-service
spec:
  selector:
    app: {{.Name}}
  ports:
  {{- range $i, $port := .ServicePorts }}
      - port: {{ $port.Port }}
        name: {{ $port.Name }}
        targetPort: {{ $port.TargetPort }}
        protocol: TCP
    {{- end }}
{{- end}}
{{ if .Dns }}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{.Name}}-ingress
  namespace: tooling
  annotations:
    kubernetes.io/ingress.class: 'nginx'
spec:
  tls:
    - hosts:
      - {{.Dns}}
      secretName: sgdev-tls-secret
  rules:
    - host: {{.Dns}}
      http:
        paths:
          - backend:
              service:
                name: {{ .Name }}-service
                port:
                  number: {{ (index .ServicePorts 0).Port }}
            path: /
            pathType: Prefix
{{- end }}
`

func generateManifest(configFile string) error {

	err := checkCurrentDir("infrastructure")
	if err != nil {
		return errors.Wrap(err, "check current directory")
	}

	var values Values
	v, err := os.ReadFile(configFile)
	if err != nil {
		return errors.Wrap(err, "read values file")
	}

	err = yaml.Unmarshal(v, &values)
	if err != nil {
		return errors.Wrap(err, "error rendering values")
	}

	path := "dogfood/kubernetes/tooling/" + values.Name
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return errors.Wrap(err, "create directory")
	}

	file, err := os.Create(path + values.Name + ".yaml")
	if err != nil {
		return errors.Wrap(err, "create file")
	}
	defer file.Close()

	t := template.Must(template.New("app").Parse(k8sTemplate))
	err = t.Execute(file, &values)
	if err != nil {
		return errors.Wrap(err, "execute template")
	}
	return nil
}

func checkCurrentDir(expected string) error {

	fmt.Println("Checking current directory")
	cwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "error getting current directory")
	}

	current := path.Base(cwd)
	if current != expected {
		return errors.Wrap(err, "Incorrect direcotory. Please run from the sourcegraph/infrastructure repository")
	}

	return nil
}
