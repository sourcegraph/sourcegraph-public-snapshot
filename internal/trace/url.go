package trace

// URL returns a trace URL for the given trace ID using the configured trace URL
// renderer. See trace.Init for more information.
func URL(traceID string) string {
	if traceID == "" {
		return ""
	}
	if urlRenderer == nil {
		return "<url renderer not configured see trace.Init>"
	}

	return urlRenderer(traceID)
}
