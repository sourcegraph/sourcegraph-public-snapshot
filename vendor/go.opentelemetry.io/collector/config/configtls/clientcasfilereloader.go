// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package configtls // import "go.opentelemetry.io/collector/config/configtls"

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type clientCAsFileReloader struct {
	clientCAsFile   string
	certPool        *x509.CertPool
	lastReloadError error
	lock            sync.RWMutex
	loader          clientCAsFileLoader
	watcher         *fsnotify.Watcher
	shutdownCH      chan bool
}

type clientCAsFileLoader interface {
	loadClientCAFile() (*x509.CertPool, error)
}

func newClientCAsReloader(clientCAsFile string, loader clientCAsFileLoader) (*clientCAsFileReloader, error) {
	certPool, err := loader.loadClientCAFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load client CA CertPool: %w", err)
	}

	reloader := &clientCAsFileReloader{
		clientCAsFile: clientCAsFile,
		certPool:      certPool,
		loader:        loader,
		shutdownCH:    nil,
		watcher:       nil,
	}

	return reloader, nil
}

func (r *clientCAsFileReloader) getClientConfig(original *tls.Config) (*tls.Config, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return &tls.Config{
		RootCAs:              original.RootCAs,
		GetCertificate:       original.GetCertificate,
		GetClientCertificate: original.GetClientCertificate,
		MinVersion:           original.MinVersion,
		MaxVersion:           original.MaxVersion,
		NextProtos:           original.NextProtos,
		ClientCAs:            r.certPool,
		ClientAuth:           tls.RequireAndVerifyClientCert,
	}, nil
}

func (r *clientCAsFileReloader) reload() {
	r.lock.Lock()
	defer r.lock.Unlock()
	certPool, err := r.loader.loadClientCAFile()
	if err != nil {
		r.lastReloadError = err
	} else {
		r.certPool = certPool
		r.lastReloadError = nil
	}
}

func (r *clientCAsFileReloader) getLastError() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.lastReloadError
}

func (r *clientCAsFileReloader) startWatching() error {
	if r.shutdownCH != nil {
		return fmt.Errorf("client CA file watcher already started")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher to reload client CA CertPool: %w", err)
	}
	r.watcher = watcher

	err = watcher.Add(r.clientCAsFile)
	if err != nil {
		return fmt.Errorf("failed to add client CA file to watcher: %w", err)
	}

	r.shutdownCH = make(chan bool)
	go r.handleWatcherEvents()

	return nil
}

func (r *clientCAsFileReloader) handleWatcherEvents() {
	defer r.watcher.Close()
	for {
		select {
		case _, ok := <-r.shutdownCH:
			_ = ok
			return
		case event, ok := <-r.watcher.Events:
			if !ok {
				continue
			}
			// NOTE: k8s configmaps uses symlinks, we need this workaround.
			// original configmap file is removed.
			// SEE: https://martensson.io/go-fsnotify-and-kubernetes-configmaps/
			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Chmod) {
				// remove the watcher since the file is removed
				if err := r.watcher.Remove(event.Name); err != nil {
					r.lastReloadError = err
				}
				// add a new watcher pointing to the new symlink/file
				if err := r.watcher.Add(r.clientCAsFile); err != nil {
					r.lastReloadError = err
				}
				r.reload()
			}
			if event.Has(fsnotify.Write) {
				r.reload()
			}
		}
	}
}

func (r *clientCAsFileReloader) shutdown() error {
	if r.shutdownCH == nil {
		return fmt.Errorf("client CAs file watcher is not running")
	}
	r.shutdownCH <- true
	close(r.shutdownCH)
	r.shutdownCH = nil
	return nil
}
