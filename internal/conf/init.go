package conf

// Init function completes the initialization process of the conf package, starting the configuration continuous changes polling
// if in client mode. The conf.Watch function can safely be called before calling Init to register callbacks reacting to the changes.
//
// The Init function must be called early in an application initialization process, but tests do not need to call it.
func Init() {
	// The default client is started in InitConfigurationServerFrontendOnly in
	// the case of server mode.
	if getMode() == modeClient {
		go DefaultClient().continuouslyUpdate(nil)
		close(configurationServerFrontendOnlyInitialized)
	}
}
