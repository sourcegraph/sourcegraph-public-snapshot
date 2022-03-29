package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractCommitFromDiff(t *testing.T) {
	diffBody := `diff --git a/base/frontend/sourcegraph-frontend-internal.Deployment.yaml b/base/frontend/sourcegraph-frontend-internal.Deployment.yaml
index 686a43f17..3b661041d 100644
--- a/base/frontend/sourcegraph-frontend-internal.Deployment.yaml
+++ b/base/frontend/sourcegraph-frontend-internal.Deployment.yaml
@@ -33,7 +33,7 @@ spec:
       initContainers:
       - name: migrator
         args: ["up", "-db=frontend,codeintel"]
-        image: index.docker.io/sourcegraph/migrator:135601_2022-03-08_7d8d9af5224d@sha256:76b48934a574b5865d0a46d5dc457376f7aee69a4dfd744d35eb65d4efd41a04
+        image: index.docker.io/sourcegraph/migrator:135692_2022-03-08_73b7bfcf2c40@sha256:2fe3c0c0aab10a4d3c82fe1468ec70251f5d2163bf6633fa24d0e663a4109e7f
         env:
           - name: PGHOST
             value: cloud-sql-proxy
@@ -46,7 +46,7 @@ spec:
             name: frontend-secrets
       containers:
       - name: frontend
-        image: index.docker.io/sourcegraph/frontend:135601_2022-03-08_7d8d9af5224d@sha256:06333ab3d44f822643b600ba1ff888e4544f9000068a3ea828e88bcdccc397e4
+        image: index.docker.io/sourcegraph/frontend:135692_2022-03-08_73b7bfcf2c40@sha256:75ed80b339e4c1b8055fed33141e9985d3eaf067fb2d477e26fad1d48330f0bb
         args:
         - serve
         envFrom:
`

	diff := parseSourcegraphCommitFromDeploymentManifestsDiff([]byte(diffBody))
	assert.Equal(t, "73b7bfcf2c40", diff.New)
	assert.Equal(t, "7d8d9af5224d", diff.Old)
}
