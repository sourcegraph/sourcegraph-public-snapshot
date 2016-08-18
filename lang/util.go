package lang

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// PrepareKeys copies SSH keys to /root/.ssh where our configs are located.
func PrepareKeys() {
	// HACK: copy files from /config to /root/.ssh because that is where our
	// configs are located. "_" (underscore) is not a valid key name in a
	// Kubernetes configmap, so we must do this or not use config maps. See:
	//
	// https://github.com/kubernetes/kubernetes/issues/13357#issuecomment-136554256
	// https://github.com/kubernetes/kubernetes/issues/16786#issue-115047222
	// https://github.com/kubernetes/kubernetes/issues/4789
	//
	move := map[string]string{
		"/config/idrsa":      "/root/.ssh/id_rsa",
		"/config/idrsa.pub":  "/root/.ssh/id_rsa.pub",
		"/config/knownhosts": "/root/.ssh/known_hosts",
	}
	for src, dst := range move {
		if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
			log.Println(err)
			continue
		}
		// We must copy the files because they live on separate devices and
		// renaming just gets us an "invalid cross-device link" error.
		srcFile, err := os.Open(src)
		if err != nil {
			log.Println(err)
			continue
		}
		dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
		if err != nil {
			log.Println(err)
			continue
		}
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			log.Println(err)
			continue
		}
		srcFile.Close()
		dstFile.Close()
	}
}
