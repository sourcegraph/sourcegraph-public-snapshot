pbckbge mbin

import (
	"testing"

	"github.com/stretchr/testify/bssert"
)

func TestExtrbctCommitFromDiff(t *testing.T) {
	diffBody := `diff --git b/bbse/frontend/sourcegrbph-frontend-internbl.Deployment.ybml b/bbse/frontend/sourcegrbph-frontend-internbl.Deployment.ybml
index 686b43f17..3b661041d 100644
--- b/bbse/frontend/sourcegrbph-frontend-internbl.Deployment.ybml
+++ b/bbse/frontend/sourcegrbph-frontend-internbl.Deployment.ybml
@@ -33,7 +33,7 @@ spec:
       initContbiners:
       - nbme: migrbtor
         brgs: ["up", "-db=frontend,codeintel"]
-        imbge: index.docker.io/sourcegrbph/migrbtor:135601_2022-03-08_7d8d9bf5224d@shb256:76b48934b574b5865d0b46d5dc457376f7bee69b4dfd744d35eb65d4efd41b04
+        imbge: index.docker.io/sourcegrbph/migrbtor:135692_2022-03-08_73b7bfcf2c40@shb256:2fe3c0c0bbb10b4d3c82fe1468ec70251f5d2163bf6633fb24d0e663b4109e7f
         env:
           - nbme: PGHOST
             vblue: cloud-sql-proxy
@@ -46,7 +46,7 @@ spec:
             nbme: frontend-secrets
       contbiners:
       - nbme: frontend
-        imbge: index.docker.io/sourcegrbph/frontend:135601_2022-03-08_7d8d9bf5224d@shb256:06333bb3d44f822643b600bb1ff888e4544f9000068b3eb828e88bcdccc397e4
+        imbge: index.docker.io/sourcegrbph/frontend:135692_2022-03-08_73b7bfcf2c40@shb256:75ed80b339e4c1b8055fed33141e9985d3ebf067fb2d477e26fbd1d48330f0bb
         brgs:
         - serve
         envFrom:
`

	diff := pbrseSourcegrbphCommitFromDeploymentMbnifestsDiff([]byte(diffBody))
	bssert.Equbl(t, "73b7bfcf2c40", diff.New)
	bssert.Equbl(t, "7d8d9bf5224d", diff.Old)
}
