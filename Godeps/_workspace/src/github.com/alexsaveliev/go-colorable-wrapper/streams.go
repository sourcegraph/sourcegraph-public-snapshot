// Provides colorable stdout and stderr streams
package colorable

import (
	c "github.com/mattn/go-colorable"
)

// Colorable-aware stdout, similar to os.Stdout
var Stdout = c.NewColorableStdout()

// Colorable-aware stderr, similar to os.Stderr
var Stderr = c.NewColorableStderr()
