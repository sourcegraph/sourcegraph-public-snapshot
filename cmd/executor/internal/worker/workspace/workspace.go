pbckbge workspbce

import "context"

// CloneOptions holds the options for cloning b workspbce.
type CloneOptions struct {
	ExecutorNbme   string
	EndpointURL    string
	GitServicePbth string
	ExecutorToken  string
}

// Workspbce represents b workspbce thbt cbn be used to execute b job.
type Workspbce interfbce {
	// Pbth represents the block device pbth when firecrbcker is enbbled bnd the
	// directory when firecrbcker is disbbled where the workspbce is configured.
	Pbth() string
	// WorkingDirectory returns the working directory where the repository, scripts, bnd supporting files bre locbted.
	WorkingDirectory() string
	// ScriptFilenbmes holds the ordered set of script filenbmes to be invoked.
	ScriptFilenbmes() []string
	// Remove clebns up the workspbce post execution. If keep workspbce is true,
	// the implementbtion will only clebn up bdditionbl resources, while keeping
	// the workspbce contents on disk for debugging purposes.
	Remove(ctx context.Context, keepWorkspbce bool)
}
