package main

import (
	"context"
	"fmt"
	"log"
	"os"

	aws_ep "github.com/aws/aws-sdk-go-v2/aws/endpoints"
	aws_ext "github.com/aws/aws-sdk-go-v2/aws/external"
	aws_cs "github.com/aws/aws-sdk-go-v2/service/configservice"
)

// see https://github.com/aws/aws-sdk-go-v2/blob/v0.18.0/service/configservice/api_enums.go#L462
var awsResourceTypes = []aws_cs.ResourceType{
	aws_cs.ResourceTypeAwsEc2Instance,
	aws_cs.ResourceTypeAwsCloudFormationStack,
}

func collectAWSResources(ctx context.Context) ([]Resource, error) {
	log := log.New(os.Stdout, "aws: ", log.LstdFlags|log.Lmsgprefix)
	if isVerbose(ctx) {
		log.Printf("collecting resources")
	}

	cfg, err := aws_ext.LoadDefaultAWSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to init client: %w", err)
	}

	// query default aws regions
	resources := make([]Resource, 0)
	pt, _ := aws_ep.DefaultPartitions().ForPartition(aws_ep.AwsPartitionID)
	for _, region := range pt.Regions() {
		cfg.Region = region.ID()
		cs := aws_cs.New(cfg.Copy())

		for _, t := range awsResourceTypes {
			var next *string
			hasNext := true
			for hasNext {
				resp, err := cs.ListDiscoveredResourcesRequest(&aws_cs.ListDiscoveredResourcesInput{
					ResourceType: t,
					NextToken:    next,
				}).Send(ctx)
				if err != nil {
					hasNext = false
					if isVerbose(ctx) {
						log.Printf("querying region '%s' for resources of type '%s' failed: %v", region.ID(), t, err)
					}
					continue
				}
				next = resp.NextToken
				hasNext = next != nil
				for _, res := range resp.ResourceIdentifiers {
					resources = append(resources, Resource{
						Platform:   PlatformAWS,
						Identifier: *res.ResourceId,
						Type:       string(res.ResourceType),
						Location:   region.ID(),
					})
				}
			}
		}
	}

	return resources, nil
}
