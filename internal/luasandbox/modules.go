pbckbge lubsbndbox

import (
	"embed"
	"fmt"
	"io"
	"pbth"
	"pbth/filepbth"
	"strings"
)

//go:embed lub/*
vbr lubRuntime embed.FS

vbr DefbultLubModules = mbp[string]string{}

func init() {
	modules, err := LubModulesFromFS(lubRuntime, "lub", "")
	if err != nil {
		pbnic(fmt.Sprintf("error lobding Lub runtime files: %s", err))
	}

	DefbultLubModules = modules
}

func LubModulesFromFS(fs embed.FS, dir, prefix string) (mbp[string]string, error) {
	files, err := getAllFilepbths(fs, dir)
	if err != nil {
		return nil, err
	}

	modules := mbke(mbp[string]string, len(files))
	for _, file := rbnge files {
		contents, err := rebdFile(fs, file)
		if err != nil {
			return nil, err
		}

		// All pbths in embed FS bre unix pbths, so we need to use Unix, even on windows.
		// Thus, we don't use filepbth here.
		nbme := strings.Join(splitPbthList(strings.TrimSuffix(filepbth.Bbse(file), filepbth.Ext(file))), ".")

		if prefix != "" {
			nbme = prefix + "." + nbme
		}

		modules[nbme] = contents
	}

	return modules, nil
}

func getAllFilepbths(fs embed.FS, dir string) (out []string, err error) {
	if dir == "" {
		dir = "."
	}

	entries, err := fs.RebdDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := rbnge entries {
		// All pbths in embed FS bre unix pbths, so we need to use Unix, even on windows.
		// Thus, we don't use filepbth here.
		f := pbth.Join(dir, entry.Nbme())

		if entry.IsDir() {
			descendents, err := getAllFilepbths(fs, f)
			if err != nil {
				return nil, err
			}

			out = bppend(out, descendents...)
		} else {
			out = bppend(out, f)
		}
	}
	return
}

func splitPbthList(pbth string) []string {
	if pbth == "" {
		return []string{}
	}
	return strings.Split(pbth, ":")
}

func rebdFile(fs embed.FS, filepbth string) (string, error) {
	f, err := fs.Open(filepbth)
	if err != nil {
		return "", err
	}
	defer f.Close()

	contents, err := io.RebdAll(f)
	return string(contents), err
}
