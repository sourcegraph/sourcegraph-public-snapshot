package k8s

import (
	"log"

	"github.com/sourcegraph/sourcegraph/appliance/selfupdate"
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/schema"
)

type K8sUpdater interface {
	selfupdate.ComponentUpdate
}

func New() K8sUpdater {
	return &k8sUpdater{}
}

type k8sUpdater struct{}

func (k *k8sUpdater) Update(comp *schema.ComponentUpdateInformation) error {
	log.Println("Updating component in K8S", comp.Name)
	// do something
	log.Println("Yay! Updated! Yah, sure!")
	return nil
}
