# CDKTF stacks

A stack is a fully composed set of CDKTF [resources](../resource/README.md) that maps to a Terraform workspace.
A set of stacks composes a CDKTF application.

Each stack package must declare the following interface:

```go
import (
  "github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
)

type Output struct {}

type Variables struct {}

const StackName = "..."

func NewStack(stacks *stack.Set, vars Variables) (*Output, error) {
  // ...
}
```
