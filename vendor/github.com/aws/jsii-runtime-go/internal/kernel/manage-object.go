package kernel

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/jsii-runtime-go/internal/api"
)

const objectFQN = "Object"

func (c *Client) ManageObject(v reflect.Value) (ref api.ObjectRef, err error) {
	// Ensuring we use a pointer, so we can see pointer-receiver methods, too.
	var vt reflect.Type
	if v.Kind() == reflect.Interface || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Interface) {
		vt = reflect.Indirect(reflect.ValueOf(v.Interface())).Addr().Type()
	} else {
		vt = reflect.Indirect(v).Addr().Type()
	}
	interfaces, overrides := c.Types().DiscoverImplementation(vt)

	found := make(map[string]bool)
	for _, override := range overrides {
		if prop, ok := override.(*api.PropertyOverride); ok {
			found[prop.JsiiProperty] = true
		}
	}
	overrides = appendExportedProperties(vt, overrides, found)

	var resp CreateResponse
	resp, err = c.Create(CreateProps{
		FQN:        objectFQN,
		Interfaces: interfaces,
		Overrides:  overrides,
	})

	if err == nil {
		if err = c.objects.Register(v, api.ObjectRef{InstanceID: resp.InstanceID, Interfaces: interfaces}); err == nil {
			ref.InstanceID = resp.InstanceID
		}
	}

	return
}

func appendExportedProperties(vt reflect.Type, overrides []api.Override, found map[string]bool) []api.Override {
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}

	if vt.Kind() == reflect.Struct {
		for idx := 0; idx < vt.NumField(); idx++ {
			field := vt.Field(idx)
			// Unexported fields are not relevant here...
			if !field.IsExported() {
				continue
			}

			// Anonymous fields are embed, we traverse them for fields, too...
			if field.Anonymous {
				overrides = appendExportedProperties(field.Type, overrides, found)
				continue
			}

			jsonName := field.Tag.Get("json")
			if jsonName == "-" {
				// Explicit omit via `json:"-"`
				continue
			} else if jsonName != "" {
				// There could be attributes after the field name (e.g. `json:"foo,omitempty"`)
				jsonName = strings.Split(jsonName, ",")[0]
			}
			// The default behavior is to use the field name as-is in JSON.
			if jsonName == "" {
				jsonName = field.Name
			}

			if !found[jsonName] {
				overrides = append(overrides, &api.PropertyOverride{
					JsiiProperty: jsonName,
					// Using the "." prefix to signify this isn't actually a getter, just raw field access.
					GoGetter: fmt.Sprintf(".%s", field.Name),
				})
				found[jsonName] = true
			}
		}
	}

	return overrides
}
