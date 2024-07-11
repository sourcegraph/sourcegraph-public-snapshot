package config

func (sg *Sourcegraph) SetLocalDevMode() {
	sg.Spec.Blobstore.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.Blobstore.ContainerConfig, "blobstore")
	sg.Spec.Cadvisor.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.Cadvisor.ContainerConfig, "cadvisor")

	sg.Spec.CodeInsights.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.CodeInsights.ContainerConfig, "correct-data-dir-permissions")
	sg.Spec.CodeInsights.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.CodeInsights.ContainerConfig, "codeinsights")
	sg.Spec.CodeInsights.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.CodeInsights.ContainerConfig, "pgsql-exporter")

	sg.Spec.CodeIntel.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.CodeIntel.ContainerConfig, "correct-data-dir-permissions")
	sg.Spec.CodeIntel.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.CodeIntel.ContainerConfig, "codeintel-db")
	sg.Spec.CodeIntel.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.CodeIntel.ContainerConfig, "pgsql-exporter")

	sg.Spec.Frontend.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.Frontend.ContainerConfig, "frontend")
	sg.Spec.Frontend.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.Frontend.ContainerConfig, "migrator")

	sg.Spec.GitServer.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.GitServer.ContainerConfig, "gitserver")
	sg.Spec.Grafana.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.Grafana.ContainerConfig, "grafana")

	sg.Spec.IndexedSearch.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.IndexedSearch.ContainerConfig, "zoekt-webserver")
	sg.Spec.IndexedSearch.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.IndexedSearch.ContainerConfig, "zoekt-indexserver")

	sg.Spec.PGSQL.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.PGSQL.ContainerConfig, "correct-data-dir-permissions")
	sg.Spec.PGSQL.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.PGSQL.ContainerConfig, "pgsql")
	sg.Spec.PGSQL.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.PGSQL.ContainerConfig, "pgsql-exporter")

	sg.Spec.PreciseCodeIntel.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.PreciseCodeIntel.ContainerConfig, "precise-code-intel-worker")
	sg.Spec.Prometheus.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.Prometheus.ContainerConfig, "prometheus")
	sg.Spec.RepoUpdater.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.RepoUpdater.ContainerConfig, "repo-updater")

	sg.Spec.RedisCache.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.RedisCache.ContainerConfig, "redis-cache")
	sg.Spec.RedisCache.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.RedisCache.ContainerConfig, "redis-exporter")
	sg.Spec.RedisStore.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.RedisStore.ContainerConfig, "redis-store")
	sg.Spec.RedisStore.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.RedisStore.ContainerConfig, "redis-exporter")

	sg.Spec.Symbols.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.Symbols.ContainerConfig, "symbols")
	sg.Spec.SyntectServer.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.SyntectServer.ContainerConfig, "syntect-server")
	sg.Spec.Worker.ContainerConfig = setBestEffortQOSOnContainer(sg.Spec.Worker.ContainerConfig, "worker")
}

func setBestEffortQOSOnContainer(ctrConfigs map[string]ContainerConfig, container string) map[string]ContainerConfig {
	if ctrConfigs == nil {
		ctrConfigs = map[string]ContainerConfig{}
	}
	ctrConfig := ctrConfigs[container]
	ctrConfig.BestEffortQOS = true
	ctrConfigs[container] = ctrConfig
	return ctrConfigs
}
