package dev

// This file contains imports that are used for development tasks
// (e.g., running the dev server). It underscore-imports packages so
// that `govendor` knows they are in use.

import (
	_ "sourcegraph.com/sqs/rego"
)
