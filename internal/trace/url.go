package trace

// URL returns a trace URL for the given trace ID using the configured trace URL
// renderer. See trace.Init for more information.
func URL(traceID string) string {
	if traceID == "" {
		return ""
	}
	urlRendererMu.Lock()
	defer urlRendererMu.Unlock()
	if urlRenderer == nil {
		return "<internal/trace.urlRenderer not configured>"
	}

	return urlRenderer(traceID)
}
