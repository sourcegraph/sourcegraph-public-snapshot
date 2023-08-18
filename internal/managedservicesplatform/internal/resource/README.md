# CDKTF resources

Resources are composable units of CDKTF resources.
A set of resources composes a CDKTF [stack](../stack/README.md).

Each resource package must declare the following interface:

```go
import (
  "github.com/aws/constructs-go/constructs/v10"
)

type Output struct {}

type Config struct {}

func New(scope constructs.Construct, id string, config Config) (*Output, error) {
  // ...
}
```
