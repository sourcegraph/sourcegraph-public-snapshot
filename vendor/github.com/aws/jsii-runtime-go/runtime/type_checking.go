package runtime

import "github.com/aws/jsii-runtime-go/internal/kernel"

// ValidateStruct runs validations on the supplied struct to determine whether
// it is valid. In particular, it checks union-typed properties to ensure the
// provided value is of one of the allowed types.
func ValidateStruct(v interface{}, d func() string) error {
	client := kernel.GetClient()
	return client.Types().ValidateStruct(v, d)
}
