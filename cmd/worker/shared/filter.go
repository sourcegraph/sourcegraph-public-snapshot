package shared

// shouldRunSetupHook returns true if the given setup hook should be run.
func shouldRunSetupHook(name string) bool {
	for _, rc := range config.SetupHookBlocklist {
		if name == rc {
			return false
		}
	}

	for _, rc := range config.SetupHookAllowlist {
		if rc == "all" || name == rc {
			return true
		}
	}

	return false
}
