package typeregistry

import (
	"reflect"
	"strings"

	"github.com/aws/jsii-runtime-go/internal/api"
)

// DiscoverImplementation determines the list of registered interfaces that are
// implemented by the provided type, and returns the list of their FQNs and
// overrides for all their combined methods and properties.
func (t *TypeRegistry) DiscoverImplementation(vt reflect.Type) (interfaces []api.FQN, overrides []api.Override) {
	if strings.HasPrefix(vt.Name(), "jsiiProxy_") {
		return
	}

	registeredOverrides := make(map[string]bool)
	embeds := t.registeredBasesOf(vt)

	pt := reflect.PtrTo(vt)

OuterLoop:
	for fqn, members := range t.typeMembers {
		iface := t.fqnToType[fqn]
		if iface.Kind == classType || !(vt.AssignableTo(iface.Type) || pt.AssignableTo(iface.Type)) {
			continue
		}
		for _, embed := range embeds {
			if embed.AssignableTo(iface.Type) {
				continue OuterLoop
			}
		}
		// Found a hit, registering it's FQN in the list!
		interfaces = append(interfaces, fqn)

		// Now, collecting all members thereof
		for _, override := range members {
			if identifier := override.GoName(); !registeredOverrides[identifier] {
				registeredOverrides[identifier] = true
				overrides = append(overrides, override)
			}
		}
	}

	return
}

// registeredBasesOf looks for known base type anonymously embedded (not
// recursively) in the given value type. Interfaces implemented by those types
// are actually not "additional interfaces" (they are implied).
func (t *TypeRegistry) registeredBasesOf(vt reflect.Type) []reflect.Type {
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	if vt.Kind() != reflect.Struct {
		return nil
	}
	n := vt.NumField()
	result := make([]reflect.Type, 0, n)
	for i := 0; i < n; i++ {
		f := vt.Field(i)
		if !f.Anonymous {
			continue
		}
		if _, ok := t.proxyMakers[f.Type]; ok {
			result = append(result, f.Type)
		}
	}
	return result
}
