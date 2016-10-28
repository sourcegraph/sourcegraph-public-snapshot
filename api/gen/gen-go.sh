#!/bin/bash

echo -ne "// GENERATED from sourcegraph.schema - DO NOT EDIT\n\npackage api\n\nvar Schema = \`" > schema.go
cat sourcegraph.schema >> schema.go
echo -ne "\`\n" >> schema.go
