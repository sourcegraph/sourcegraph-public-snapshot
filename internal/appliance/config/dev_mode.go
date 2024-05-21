package config

func SetLocalDevMode(sg *Sourcegraph) {
	sg.Spec.Blobstore.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.Blobstore.ContainerConfig, "blobstore")
	sg.Spec.GitServer.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.GitServer.ContainerConfig, "gitserver")

	sg.Spec.PGSQL.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.PGSQL.ContainerConfig, "correct-data-dir-permissions")
	sg.Spec.PGSQL.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.PGSQL.ContainerConfig, "pgsql")
	sg.Spec.PGSQL.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.PGSQL.ContainerConfig, "pgsql-exporter")

	sg.Spec.RedisCache.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.RedisCache.ContainerConfig, "redis-cache")
	sg.Spec.RedisCache.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.RedisCache.ContainerConfig, "redis-exporter")
	sg.Spec.RedisStore.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.RedisStore.ContainerConfig, "redis-store")
	sg.Spec.RedisStore.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.RedisStore.ContainerConfig, "redis-exporter")

	sg.Spec.PreciseCodeIntel.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.PreciseCodeIntel.ContainerConfig, "precise-code-intel-worker")
	sg.Spec.Symbols.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.Symbols.ContainerConfig, "symbols")
	sg.Spec.SyntectServer.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.SyntectServer.ContainerConfig, "syntect-server")
	sg.Spec.RepoUpdater.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.RepoUpdater.ContainerConfig, "repo-updater")
}

func setBestEffortQOSOnContainer(ctrConfigs map[string]ContainerConfig, container string) map[string]ContainerConfig {
	if ctrConfigs == nil {
		ctrConfigs = map[string]ContainerConfig{}
	}
	ctrConfig := ctrConfigs[container]
	ctrConfig.BestEffortQOS = true
	ctrConfig.Resources = nil
	ctrConfigs[container] = ctrConfig
	return ctrConfigs
}
