package golly

import "github.com/sourcegraph/sourcegraph/internal/env"

var (
	RecordingMode      = env.Get("GOLLY_RECORDING_MODE", "replay", "Whether to record, replay or passthrough golly traces")
	RecordingDir       = env.Get("GOLLY_RECORDING_DIR", "golly-recordings", "Directory to store golly recordings")
	DotcomAccessToken  = env.Get("GOLLY_DOTCOM_ACCESS_TOKEN", "", "Dotcom access token")
	S2AccessToken      = env.Get("GOLLY_S2_ACCESS_TOKEN", "", "Sourcegraph 2 access token")
	ForceSaveRecording = env.Get("GOLLY_FORCE_SAVE_RECORDING", "false", "Whether to save recording even if recording mode is not record")
)
