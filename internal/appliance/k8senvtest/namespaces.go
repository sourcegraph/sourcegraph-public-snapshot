package k8senvtest

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// In order to be able to run tests in isolation, we can make use of namespaces
// with a random suffix. We don't need to delete these, all data will be
// desstroyed on envtest teardown.
func NewRandomNamespace(prefix string) (*corev1.Namespace, error) {
	slug, err := randomSlug()
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("%s-%s", prefix, slug)
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}, nil
}

func randomSlug() (string, error) {
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
