package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangesetSpec_ForkGetters(t *testing.T) {
	for name, tc := range map[string]struct {
		spec      *ChangesetSpec
		isFork    bool
		namespace *string
	}{
		"no fork": {
			spec:      &ChangesetSpec{ForkNamespace: nil},
			isFork:    false,
			namespace: nil,
		},
		"fork to user": {
			spec:      &ChangesetSpec{ForkNamespace: strPtr(changesetSpecForkNamespaceUser)},
			isFork:    true,
			namespace: nil,
		},
		"fork to namespace": {
			spec:      &ChangesetSpec{ForkNamespace: strPtr("org")},
			isFork:    true,
			namespace: strPtr("org"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.isFork, tc.spec.IsFork())
			if tc.namespace == nil {
				assert.Nil(t, tc.spec.GetForkNamespace())
			} else {
				have := tc.spec.GetForkNamespace()
				assert.NotNil(t, have)
				assert.Equal(t, *tc.namespace, *have)
			}
		})
	}
}

func TestChangesetSpec_SetForkToUser(t *testing.T) {
	cs := &ChangesetSpec{ForkNamespace: nil}
	cs.setForkToUser()
	assert.NotNil(t, cs.ForkNamespace)
	assert.Equal(t, changesetSpecForkNamespaceUser, *cs.ForkNamespace)
}

func strPtr(s string) *string { return &s }
