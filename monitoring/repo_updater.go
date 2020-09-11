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
						sharedFrontendInternalAPIErrorResponses("repo-updater", ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerCPUUsage("repo-updater", ObservableOwnerCloud),
						sharedContainerMemoryUsage("repo-updater", ObservableOwnerCloud),
					},
					{
						sharedContainerRestarts("repo-updater", ObservableOwnerCloud),
						sharedContainerFsInodes("repo-updater", ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsageLongTerm("repo-updater", ObservableOwnerCloud),
						sharedProvisioningMemoryUsageLongTerm("repo-updater", ObservableOwnerCloud),
					},
					{
						sharedProvisioningCPUUsageShortTerm("repo-updater", ObservableOwnerCloud),
						sharedProvisioningMemoryUsageShortTerm("repo-updater", ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []Row{
					{
						sharedGoGoroutines("repo-updater", ObservableOwnerCloud),
						sharedGoGcDuration("repo-updater", ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedKubernetesPodsAvailable("repo-updater", ObservableOwnerCloud),
					},
				},
			},
		},
	}
}
