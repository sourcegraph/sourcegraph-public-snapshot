package main

func RepoUpdater() *Container {
	return &Container{
		Name:        "repo-updater",
		Title:       "Repo Updater",
		Description: "Manages interaction with code hosts, instructs Gitserver to update repositories.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						sharedFrontendInternalAPIErrorResponses("repo-updater"),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerCPUUsage("repo-updater"),
						sharedContainerMemoryUsage("repo-updater"),
					},
					{
						sharedContainerRestarts("repo-updater"),
						sharedContainerFsInodes("repo-updater"),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsage7d("repo-updater"),
						sharedProvisioningMemoryUsage7d("repo-updater"),
					},
					{
						sharedProvisioningCPUUsage5m("repo-updater"),
						sharedProvisioningMemoryUsage5m("repo-updater"),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []Row{
					{
						sharedGoGoroutines("repo-updater"),
						sharedGoGcDuration("repo-updater"),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedKubernetesPodsAvailable("repo-updater"),
					},
				},
			},
		},
	}
}
