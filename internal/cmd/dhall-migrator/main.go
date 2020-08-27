package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"gopkg.in/yaml.v3"
)

type Resource struct {
	Source     string
	Component  string
	Kind       string
	ApiVersion string
	Name       string
	DhallType  string
	Labels     map[string]string
	Contents   map[string]interface{}
}

type ResourceSet struct {
	Root       string
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

	res.DhallType = fmt.Sprintf("(https://raw.githubusercontent.com/dhall-lang/dhall-kubernetes/f4bf4b9ddf669f7149ec32150863a93d6c4b3ef1/1.18/schemas.dhall).%s.Type", res.Kind)

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
			rs.Components[res.Component] = append(rs.Components[res.Component], res)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &rs, nil
}

func composeDhallSchema(rs *ResourceSet) string {
	var schemas []string

	for component, crs := range rs.Components {
		for _, res := range crs {
			s := fmt.Sprintf("{%s: { %s: { %s: %s } } }", component, res.Kind, res.Name, res.DhallType)
			schemas = append(schemas, s)
		}
	}
	return strings.Join(schemas, " //\\\\ ")
}

func buildRecord(rs *ResourceSet) map[string]interface{} {
	rec := make(map[string]interface{})
	for component, crs := range rs.Components {
		compRec := make(map[string]map[string]interface{})
		rec[strings.Title(component)] = compRec
		for _, res := range crs {
			kindRec := compRec[res.Kind]
			if kindRec == nil {
				kindRec = make(map[string]interface{})
				compRec[res.Kind] = kindRec
			}
			kindRec[res.Name] = res.Contents
		}
	}
	return rec
}

func writeRecordYaml(filename string, rec map[string]interface{}) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	bw := bufio.NewWriter(f)
	defer bw.Flush()

	enc := yaml.NewEncoder(bw)
	return enc.Encode(rec)
}

func execYamlToDhall(scratchDir string, dst string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yaml-to-dhall", "./schema.dhall", "--records-loose", "--file",
		"record.yaml", "--output", dst)
	cmd.Dir = scratchDir
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func main() {
	src := flag.String("src", "", "(required) source manifest directory")
	dst := flag.String("dst", "", "(required) output dhall file")

	scratchDir := flag.String("scratchDir", "", "scratch dir used in migration")

	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	log15.Root().SetHandler(log15.StreamHandler(os.Stdout, log15.LogfmtFormat()))

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

	dRec := buildRecord(srcSet)
	yamlFilename := filepath.Join(*scratchDir, "record.yaml")

	log15.Info("building yaml record", "yamlFile", yamlFilename)

	err = writeRecordYaml(yamlFilename, dRec)
	if err != nil {
		log15.Error("failed to write out record yaml file", "error", err, "yamlFile", yamlFilename)
		os.Exit(1)
	}

	log15.Info("execute yaml-to-dhall", "scratchDir", *scratchDir, "dst", *dst)

	err = execYamlToDhall(*scratchDir, *dst)
	if err != nil {
		log15.Error("failed to execute yaml-to-dhall", "error", err, "scratchDir", *scratchDir, "dst", *dst)
		os.Exit(1)
	}

	log15.Info("done")
}
