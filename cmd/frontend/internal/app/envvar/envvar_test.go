package envvar

import "testing"

func TestDeploymentOnPrem(t *testing.T) {
	deploymentOnPrem = false
	if DeploymentOnPrem() {
		t.Fatalf("expected DeploymentOnPrem() = false; got true")
	}
	deploymentOnPrem = true
	if !DeploymentOnPrem() {
		t.Fatalf("expected DeploymentOnPrem() = true; got false")
	}
}
