package runner

import "context"

type kubernetesRunner struct {
}

var _ Runner = &kubernetesRunner{}

func (k *kubernetesRunner) Setup(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (k *kubernetesRunner) TempDir() string {
	//TODO implement me
	panic("implement me")
}

func (k *kubernetesRunner) Teardown(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (k *kubernetesRunner) Run(ctx context.Context, spec Spec) error {
	//TODO implement me
	panic("implement me")
}
