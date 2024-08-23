package typeregistry

import (
	"github.com/aws/jsii-runtime-go/internal/api"
)

func (t *TypeRegistry) GetOverride(fqn api.FQN, n string) (api.Override, bool) {
	if members, ok := t.typeMembers[fqn]; ok {
		for _, member := range members {
			if member.GoName() == n {
				return member, true
			}
		}
	}

	return nil, false
}
