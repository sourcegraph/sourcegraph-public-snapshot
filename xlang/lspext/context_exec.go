package lspext

// ExecParams contains the parameters for the exec LSP request.
type ExecParams struct {
	Command   string   `json:"command"`
	Arguments []string `json:"arguments"`
}

// ExecResult contains the result for the exec LSP response.
type ExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}
