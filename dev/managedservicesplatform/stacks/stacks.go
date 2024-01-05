package stacks

import "fmt"

// OutputSecretID scopes the secret ID storing the value of a stack variable
// under the stack name to avoid conflicts across stacks.
//
// See stack.StackLocals.
func OutputSecretID(stackName, varName string) string {
	return fmt.Sprintf("%s_%s", stackName, varName)
}
