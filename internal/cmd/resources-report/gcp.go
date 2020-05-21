package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	gcp_crm "google.golang.org/api/cloudresourcemanager/v1"
	gcp_cp "google.golang.org/api/compute/v1"
	gcp_ct "google.golang.org/api/container/v1"
)

// for resources that require enumerating over regions, it is not very partical to
// make queries for regions that will either not work or will never have Sourcegraph
// resources - define relevant regions here.
//
// https://cloud.google.com/compute/docs/regions-zones#locations
var gcpLocationPrefixes = []string{
	"us-",
	"europe-",
}

type GCPResourceFetchFunc func(context.Context, string, time.Time) ([]Resource, error)

var gcpResources = map[string]GCPResourceFetchFunc{
	// fetch disks and instances
	"compute": func(ctx context.Context, project string, since time.Time) ([]Resource, error) {
		client, err := gcp_cp.NewService(ctx)
		if err != nil {
			return nil, err
		}

		// compute APIs require us to specify zones, so we must iterate over all zones
		var rs []Resource
		if err := client.Zones.List(project).Pages(ctx, func(zones *gcp_cp.ZoneList) error {
			for _, zone := range zones.Items {
				if !hasPrefix(zone.Name, gcpLocationPrefixes) {
					continue // skip this zone
				}

				if err := client.Instances.List(project, zone.Name).
					Pages(ctx, func(instances *gcp_cp.InstanceList) error {
						for _, instance := range instances.Items {
							t, err := time.Parse(time.RFC3339, instance.CreationTimestamp)
							if err != nil {
								return err
							}
							if t.After(since) {
								rs = append(rs, Resource{
									Platform:   PlatformGCP,
									Identifier: instance.Name,
									Location:   zone.Name,
									Owner:      project,
									Type:       instance.Kind,
									Meta: map[string]interface{}{
										"labels": instance.Labels,
									},
								})
							}
						}
						return nil
					}); err != nil {
					return err
				}

				if err := client.Disks.List(project, zone.Name).
					Pages(ctx, func(disks *gcp_cp.DiskList) error {
						for _, disk := range disks.Items {
							t, err := time.Parse(time.RFC3339, disk.CreationTimestamp)
							if err != nil {
								return err
							}
							if t.After(since) {
								rs = append(rs, Resource{
									Platform:   PlatformGCP,
									Identifier: disk.Name,
									Location:   zone.Name,
									Owner:      project,
									Type:       disk.Kind,
									Meta: map[string]interface{}{
										"labels": disk.Labels,
									},
								})
							}
						}
						return nil
					}); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return rs, nil
	},
	// fetch clusters
	"containers": func(ctx context.Context, project string, since time.Time) ([]Resource, error) {
		client, err := gcp_ct.NewService(ctx)
		if err != nil {
			return nil, err
		}

		// cluster api allows us to query all locations at once
		parent := fmt.Sprintf("projects/%s/locations/-", project)
		list, err := client.Projects.Locations.Clusters.List(parent).Context(ctx).Do()
		if err != nil {
			return nil, err
		}
		var rs []Resource
		for _, cluster := range list.Clusters {
			t, err := time.Parse(time.RFC3339, cluster.CreateTime)
			if err != nil {
				return nil, err
			}
			if t.After(since) {
				rs = append(rs, Resource{
					Platform:   PlatformGCP,
					Identifier: cluster.Name,
					Type:       "container#cluster",
					Owner:      project,
					Location:   cluster.Zone,
					Meta: map[string]interface{}{
						"labels": cluster.ResourceLabels,
					},
				})
			}
		}
		return rs, nil
	},
}

func collectGCPResources(ctx context.Context, since time.Time, verbose bool) ([]Resource, error) {
	logger := log.New(os.Stdout, "gcp: ", log.LstdFlags|log.Lmsgprefix)
	if verbose {
		logger.Printf("collecting resources since %s", since)
	}

	crm, err := gcp_crm.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init cloud resources client: %w", err)
	}

	// aggregate resources for each GCP project
	var resources []Resource
	if err := crm.Projects.List().Pages(ctx, func(page *gcp_crm.ListProjectsResponse) error {
		for _, project := range page.Projects {
			for resourceID, fetchResource := range gcpResources {
				rs, err := fetchResource(ctx, project.ProjectId, since)
				if err != nil {
					if verbose {
						logger.Printf("resource fetch for '%s' failed in project: %v", resourceID, err)
					}
					continue
				}
				resources = append(resources, rs...)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return resources, nil
}
