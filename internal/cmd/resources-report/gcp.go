package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	gcp_ca "google.golang.org/api/cloudasset/v1p1beta1"
	gcp_crm "google.golang.org/api/cloudresourcemanager/v1"
)

// https://cloud.google.com/asset-inventory/docs/supported-asset-types#searchable_asset_types
var gcpAssetTypes = []string{
	"compute.googleapis.com/Disk",
	"compute.googleapis.com/Instance",
	"dataproc.googleapis.com/Cluster",
}

func collectGCPResources(ctx context.Context, since time.Time, verbose bool) ([]Resource, error) {
	logger := log.New(os.Stdout, "gcp: ", log.LstdFlags|log.Lmsgprefix)
	if verbose {
		logger.Printf("collecting resources with types: %+v", gcpAssetTypes)
	}

	crm, err := gcp_crm.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init cloud resources client: %w", err)
	}
	assets, err := gcp_ca.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init assets client: %w", err)
	}

	var resources []Resource
	if err := crm.Projects.List().Pages(ctx, func(page *gcp_crm.ListProjectsResponse) error {
		for _, project := range page.Projects {
			parent := fmt.Sprintf("projects/%s", project.ProjectId)
			if verbose {
				logger.Printf("found project: %s", parent)
			}

			if err := assets.Resources.SearchAll(parent).AssetTypes(gcpAssetTypes...).
				Pages(ctx, func(page *gcp_ca.SearchAllResourcesResponse) error {
					for _, asset := range page.Results {
						resources = append(resources, Resource{
							Platform:   PlatformGCP,
							Identifier: asset.DisplayName,
							Location:   asset.Location,
							Owner:      project.ProjectId,
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
