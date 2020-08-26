package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"gopkg.in/yaml.v3"
)

const (
	dhallTypesURLPrefix = "https://raw.githubusercontent.com/dhall-lang/dhall-kubernetes/f4bf4b9ddf669f7149ec32150863a93d6c4b3ef1/1.18/types/"
)

type Resource struct {
	Source              string
	Component           string
	Kind                string
	ApiVersion          string
	Name                string
	DhallTypesURLSuffix string
	Labels              map[string]string
	Contents            map[string]interface{}
}

func (res *Resource) Key() string {
	return res.Kind + "-" + res.Name
}

type ResourceSet struct {
	Root       string
	SortedKeys []string
	Resources  map[string]*Resource
	Components map[string][]*Resource
}

func loadResource(rootDir string, filename string) (*Resource, error) {
	relPath, err := filepath.Rel(rootDir, filename)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	br := bufio.NewReader(f)
	decoder := yaml.NewDecoder(br)

	var res Resource
	res.Source = filename
	// TODO(uwedeportivo): derive it from metadata labels instead once those labels become available
	res.Component = filepath.Dir(relPath)
	if res.Component == "." {
		res.Component = filepath.Base(rootDir)
	}
	err = decoder.Decode(&res.Contents)

	kind, ok := res.Contents["kind"].(string)
	if !ok {
		return nil, fmt.Errorf("resource %s is missing a kind field", filename)
	}
	res.Kind = kind

	apiVersion, ok := res.Contents["apiVersion"].(string)
	if !ok {
		return nil, fmt.Errorf("resource %s is missing a apiVersion field", filename)
	}
	res.ApiVersion = apiVersion

	switch res.ApiVersion {
	case "v1":
		res.DhallTypesURLSuffix = fmt.Sprintf("io.k8s.api.core.v1.%s.dhall", res.Kind)
	case "rbac.authorization.k8s.io/v1":
		res.DhallTypesURLSuffix = fmt.Sprintf("io.k8s.api.rbac.v1.%s.dhall", res.Kind)
	case "apps/v1":
		res.DhallTypesURLSuffix = fmt.Sprintf("io.k8s.api.apps.v1.%s.dhall", res.Kind)
	case "networking.k8s.io/v1beta1":
		res.DhallTypesURLSuffix = fmt.Sprintf("io.k8s.api.networking.v1beta1.%s.dhall", res.Kind)
	case "policy/v1beta1":
		res.DhallTypesURLSuffix = fmt.Sprintf("io.k8s.api.policy.v1beta1.%s.dhall", res.Kind)
	case "storage.k8s.io/v1":
		res.DhallTypesURLSuffix = fmt.Sprintf("io.k8s.api.storage.v1.%s.dhall\n", res.Kind)
	default:
		return nil, fmt.Errorf("resource %s has unknown api version %s and kind %s combination", filename, res.ApiVersion, res.Kind)
	}

	metadata, ok := res.Contents["metadata"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resource %s is missing metadata", filename)
	}

	name, ok := metadata["name"].(string)
	if !ok {
		return nil, fmt.Errorf("resource %s is missing name", filename)
	}
	res.Name = name

	// patch statefulsets
	if res.Kind == "StatefulSet" {
		spec, ok := res.Contents["spec"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("resource %s is missing spec", filename)
		}
		volumeClaimTemplates, ok := spec["volumeClaimTemplates"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("resource %s is missing volumeClaimTemplates", filename)
		}
		for _, volumeClaimTemplate := range volumeClaimTemplates {
			vct, ok := volumeClaimTemplate.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("resource %s is missing volumeClaimTemplate", filename)
			}
			vct["apiVersion"] = "apps/v1"
			vct["kind"] = "PersistentVolumeClaim"
		}
	}

	return &res, err
}

func loadResourceSet(dirname string) (*ResourceSet, error) {
	dir, err := filepath.Abs(dirname)
	if err != nil {
		return nil, err
	}
	var rs ResourceSet
	rs.Resources = make(map[string]*Resource)
	rs.Components = make(map[string][]*Resource)
	rs.Root = dir

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
			res, err := loadResource(rs.Root, path)
			if err != nil {
				return err
			}
			key := res.Key()
			rs.Resources[key] = res
			rs.SortedKeys = append(rs.SortedKeys, key)
			rs.Components[res.Component] = append(rs.Components[res.Component], res)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(rs.SortedKeys)
	return &rs, nil
}

func composeDhallSchema(rs *ResourceSet) string {
	var sb strings.Builder

	sb.WriteString("{\n")

	first := true
	for component, crs := range rs.Components {
		if first {
			first = false
			sb.WriteString("\t  ")
		} else {
			sb.WriteString("\t, ")
		}
		sb.WriteString(fmt.Sprintf("%s : {\n", strings.Title(component)))
		for i, res := range crs {
			if i > 0 {
				sb.WriteString("\t\t, ")
			} else {
				sb.WriteString("\t\t  ")
			}
			sb.WriteString(fmt.Sprintf("%s_%s : %s\n", res.Kind, res.Name,
				dhallTypesURLPrefix+res.DhallTypesURLSuffix))
		}
		sb.WriteString("\t}\n")
	}
	sb.WriteString("}\n")
	return sb.String()
}

func buildDhallRecord(rs *ResourceSet) map[string]interface{} {
	drec := make(map[string]interface{})
	for component, crs := range rs.Components {
		compDRec := make(map[string]interface{})
		drec[strings.Title(component)] = compDRec
		for _, res := range crs {
			compDRec[fmt.Sprintf("%s_%s", res.Kind, res.Name)] = res.Contents
		}
	}
	return drec
}

func writeDhallRecord(filename string, dRec map[string]interface{}) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	bw := bufio.NewWriter(f)
	defer bw.Flush()

	enc := yaml.NewEncoder(bw)
	return enc.Encode(dRec)
}

func execYamlToDhall(scratchDir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	// yaml-to-dhall ./schema.dhall --records-loose --file record.yaml --output record.dhall
	cmd := exec.CommandContext(ctx, "yaml-to-dhall", "--records-loose", "--file", "record.yaml", "--output", "record.dhall")
	cmd.Dir = scratchDir

	return cmd.Run()
}

func main() {
	src := flag.String("src", "", "(required) source manifest directory")
	dst := flag.String("dst", "", "(required) output dhall file")

	logFilepath := flag.String("logfile", "dhall-migrator.log", "path to a log file")
	scratchDir := flag.String("scratchDir", "", "scratch dir used in migration")

	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	logHandler, err := log15.FileHandler(*logFilepath, log15.LogfmtFormat())
	if err != nil {
		log.Fatal(err)
	}
	log15.Root().SetHandler(log15.MultiHandler(logHandler, log15.StreamHandler(os.Stdout, log15.LogfmtFormat())))

	if *help || len(*src) == 0 || len(*dst) == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if len(*scratchDir) == 0 {
		d, err := ioutil.TempDir("", "dhall-migrator")
		if err != nil {
			log15.Error("failed to create scratch dir", "error", err)
			os.Exit(1)
		}
		*scratchDir = d
	}

	log15.Info("using scratch dir", "scratchDir", *scratchDir)

	log15.Info("loading resources", "src", *src)
	srcSet, err := loadResourceSet(*src)
	if err != nil {
		log15.Error("failed to load source resources", "error", err, "src", *src)
		os.Exit(1)
	}

	schemaFilename := filepath.Join(*scratchDir, "schema.dhall")
	log15.Info("building schema file", "schemaFile", schemaFilename)

	err = ioutil.WriteFile(schemaFilename, []byte(composeDhallSchema(srcSet)), 0777)
	if err != nil {
		log15.Error("failed to write out schema file", "error", err, "schemaFile", schemaFilename)
		os.Exit(1)
	}

	dRec := buildDhallRecord(srcSet)
	yamlFilename := filepath.Join(*scratchDir, "record.yaml")

	log15.Info("building yaml record", "yamlFile", yamlFilename)

	err = writeDhallRecord(yamlFilename, dRec)
	if err != nil {
		log15.Error("failed to write out record yaml file", "error", err, "yamlFile", yamlFilename)
		os.Exit(1)
	}

	log15.Info("execute yaml-to-dhall", "scratchDir", *scratchDir)

	err = execYamlToDhall(*scratchDir)
	if err != nil {
		log15.Error("failed to execute yaml-to-dhall", "error", err, "scratchDir", *scratchDir)
		os.Exit(1)
	}

	log15.Info("cp into destination", "dst", *dst)
	cpCmd := exec.Command("cp", filepath.Join(*scratchDir, "record.dhall"), *dst)
	err = cpCmd.Run()
	if err != nil {
		log15.Error("failed to copy record.dhall", "error", err,
			"scratchDir", *scratchDir, "destination", *dst)
		os.Exit(1)
	}
	log15.Info("done")
}
