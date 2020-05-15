package main

import (
	"context"
	"fmt"
	"log"
	"os"

	gcp_ca "google.golang.org/api/cloudasset/v1p1beta1"
	gcp_crm "google.golang.org/api/cloudresourcemanager/v1"
)

// https://cloud.google.com/asset-inventory/docs/supported-asset-types#searchable_asset_types
var gcpAssetTypes = []string{
	"appengine.googleapis.com/Version",
	"compute.googleapis.com/Disk",
	"compute.googleapis.com/Instance",
	"compute.googleapis.com/InstanceGroup",
	"dataproc.googleapis.com/Cluster",
}

func collectGCPResources(ctx context.Context) ([]Resource, error) {
	log := log.New(os.Stdout, "gcp: ", log.LstdFlags|log.Lmsgprefix)
	if isVerbose(ctx) {
		log.Printf("collecting resources with types: %+v", gcpAssetTypes)
	}

	crm, err := gcp_crm.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init cloud resources client: %w", err)
	}
	assets, err := gcp_ca.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init assets client: %w", err)
	}

	resources := make([]Resource, 0)
	if err := crm.Projects.List().Pages(ctx, func(page *gcp_crm.ListProjectsResponse) error {
		for _, p := range page.Projects {
			parent := fmt.Sprintf("projects/%s", p.ProjectId)
			if isVerbose(ctx) {
				log.Printf("found project: %s", parent)
			}

			if err := assets.Resources.SearchAll(parent).AssetTypes(gcpAssetTypes...).
				Pages(ctx, func(page *gcp_ca.SearchAllResourcesResponse) error {
					for _, asset := range page.Results {
						resources = append(resources, Resource{
							Platform:   PlatformGCP,
							Identifier: asset.Name,
							Location:   fmt.Sprintf("%s/%s", p.ProjectId, asset.Location),
							Type:       asset.AssetType,
							Meta: map[string]interface{}{
								"description": asset.Description,
								"attributes":  asset.AdditionalAttributes,
							},
						})
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
	return resources, nil
}
