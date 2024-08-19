package workspacerun


type WorkspaceRunDestroy struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/tfe/0.51.0/docs/resources/workspace_run#manual_confirm WorkspaceRun#manual_confirm}.
	ManualConfirm interface{} `field:"required" json:"manualConfirm" yaml:"manualConfirm"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/tfe/0.51.0/docs/resources/workspace_run#retry WorkspaceRun#retry}.
	Retry interface{} `field:"optional" json:"retry" yaml:"retry"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/tfe/0.51.0/docs/resources/workspace_run#retry_attempts WorkspaceRun#retry_attempts}.
	RetryAttempts *float64 `field:"optional" json:"retryAttempts" yaml:"retryAttempts"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/tfe/0.51.0/docs/resources/workspace_run#retry_backoff_max WorkspaceRun#retry_backoff_max}.
	RetryBackoffMax *float64 `field:"optional" json:"retryBackoffMax" yaml:"retryBackoffMax"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/tfe/0.51.0/docs/resources/workspace_run#retry_backoff_min WorkspaceRun#retry_backoff_min}.
	RetryBackoffMin *float64 `field:"optional" json:"retryBackoffMin" yaml:"retryBackoffMin"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/tfe/0.51.0/docs/resources/workspace_run#wait_for_run WorkspaceRun#wait_for_run}.
	WaitForRun interface{} `field:"optional" json:"waitForRun" yaml:"waitForRun"`
}

