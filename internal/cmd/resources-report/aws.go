package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_ep "github.com/aws/aws-sdk-go-v2/aws/endpoints"
	aws_rg "github.com/aws/aws-sdk-go-v2/service/resourcegroups"
)

func collectAWSResources(ctx context.Context) ([]Resource, error) {
	log := log.New(os.Stdout, "aws: ", log.LstdFlags|log.Lmsgprefix)
	if isVerbose(ctx) {
		log.Printf("collecting resources")
	}

	resources := make([]Resource, 0)
	for _, p := range aws_ep.NewDefaultResolver().Partitions() {
		for _, region := range p.Regions() {
			if isVerbose(ctx) {
				log.Printf("querying resources in region %s", region.ID())
			}
			rg := aws_rg.New(aws.Config{
				Region: region.ID(),
			})
			pager := aws_rg.NewSearchResourcesPaginator(rg.SearchResourcesRequest(&aws_rg.SearchResourcesInput{
				ResourceQuery: &aws_rg.ResourceQuery{
					// TODO
				},
			}))
			for pager.Next(ctx) {
				page := pager.CurrentPage()
				for _, r := range page.ResourceIdentifiers {
					println(*r.ResourceArn)
					resources = append(resources, Resource{
						Platform:   PlatformAWS,
						Identifier: *r.ResourceArn,
						Type:       *r.ResourceType,
						// TODO
					})
				}
			}
			if err := pager.Err(); err != nil {
				return nil, fmt.Errorf("failed to query resources in region %s: %w", region.ID(), err)
			}
		}
	}

	return resources, nil
}
