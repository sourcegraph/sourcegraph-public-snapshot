package hash

import (
	"fmt"
	"hash"
	"hash/fnv"

	"k8s.io/apimachinery/pkg/util/dump"
)

const (
	TemplateHashLabelName = "sourcegraph.k8s.sourcegraph.com/template-hash"
)

// SetTemplateHashLabel adds a label containing the hash of the given template into the provided labels.
// This label can then be used for safe comparison between templates.
func SetTemplateHashLabel(labels map[string]string, template any) map[string]string {
	return setHashLabel(TemplateHashLabelName, labels, template)
}

func setHashLabel(labelName string, labels map[string]string, template any) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}
	labels[labelName] = HashObject(template)

	return labels
}

// GetTemplateHashLabel returns the template hash label value or an empty string if value is not set.
func GetTemplateHashLabel(labels map[string]string) string {
	return labels[TemplateHashLabelName]
}

// HashObject returns the hash of a given object using the 32-bit FNV-1 hash function
// and the spew library to print the object.
// Inspired by controller revisions in StatefulSets:
// https://github.com/kubernetes/kubernetes/blob/a8f6ea24209b332217f7baa7c14248e8d0266d28/pkg/controller/history/controller_history.go#L92
func HashObject(object any) string {
	objectHash := fnv.New32a()
	WriteHash(objectHash, object)

	return fmt.Sprint(objectHash.Sum32())
}

// WriteHash writes the specified object to hash using the spew library
// which follows pointers and prints actual values of nested objects
// ensuring the hash does not change when a pointer changes.
// Copy of https://github.com/kubernetes/kubernetes/blob/a8f6ea24209b332217f7baa7c14248e8d0266d28/pkg/util/hash/hash.go#L29
func WriteHash(hasher hash.Hash, objectToWrite any) {
	hasher.Reset()
	fmt.Fprintf(hasher, "%v", dump.ForHash(objectToWrite))
}
