package graphqlbackend

import (
	"context"
	"fmt"
)

type HasPermissionsDirective struct {
	Permissions []string
}

func (h *HasPermissionsDirective) ImplementsDirective() string {
	return "hasPermissions"
}

func (h *HasPermissionsDirective) Validate(ctx context.Context, _ interface{}) error {
	// u, ok := user.FromContext(ctx)
	// if !ok {
	// 	return fmt.Errorf("user not provided in cotext")
	// }
	// role := strings.ToLower(h.Role)
	// if !u.HasRole(role) {
	// 	return fmt.Errorf("access denied, %q role required", role)
	// }
	fmt.Println("inside validate...")
	return nil
}
