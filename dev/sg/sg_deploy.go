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
		TargetPort interface{}
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
    - protocol: TCP
      port:
      targetPort:
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
  rules:
    - http:
        paths:
          - pathType: Prefix
            path: /
            backend:
              service:
                name: {{.Name}}
                port:
                  number: 80
      host: {{.Dns}}
{{- end }}
`

func generateManifest(configFile string) error {

	// err := checkCurrentDir("infrastructure")
	// if err != nil {
	// 	return errors.Wrap(err, "check current directory")
	// }

	var values Values
	v, err := os.ReadFile(configFile)
	if err != nil {
		return errors.Wrap(err, "read values file")
	}

	err = yaml.Unmarshal(v, &values)
	if err != nil {
		return errors.Wrap(err, "error rendering values")
	}
	path := "dogfood/kubernetes/tooling/" + values.Name + "/"
	file, err := os.Create(path + values.Name + ".yaml")
	if err != nil {
		panic(err)
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
		return errors.Wrap(err, "incorrect directory detected")
	}

	return nil
}
