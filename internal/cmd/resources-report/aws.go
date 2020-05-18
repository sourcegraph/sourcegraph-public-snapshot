package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_ep "github.com/aws/aws-sdk-go-v2/aws/endpoints"
	aws_ext "github.com/aws/aws-sdk-go-v2/aws/external"
	aws_ec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	aws_eks "github.com/aws/aws-sdk-go-v2/service/eks"
)

type AWSResourceFetchFunc func(context.Context, aws.Config) ([]Resource, error)

// refer to https://docs.aws.amazon.com/sdk-for-go/v2/api/ for how to query resources under /service
var awsResources = map[string]AWSResourceFetchFunc{
	// fetch ec2 instances
	"EC2::Instances": func(ctx context.Context, cfg aws.Config) ([]Resource, error) {
		c := aws_ec2.New(cfg)
		p := aws_ec2.NewDescribeInstancesPaginator(c.DescribeInstancesRequest(&aws_ec2.DescribeInstancesInput{}))
		r := make([]Resource, 0)
		for p.Next(ctx) {
			page := p.CurrentPage()
			for _, res := range page.Reservations {
				for _, inst := range res.Instances {
					r = append(r, Resource{
						Platform:   PlatformAWS,
						Identifier: *inst.InstanceId,
						Location:   *inst.Placement.AvailabilityZone,
						Owner:      *res.OwnerId,
						Type:       fmt.Sprintf("EC2::%s", string(inst.InstanceType)),
						Meta:       map[string]interface{}{},
					})
				}
			}
		}
		return r, p.Err()
	},
	// fetch kubernetes clusters
	"EKS::Clusters": func(ctx context.Context, cfg aws.Config) ([]Resource, error) {
		c := aws_eks.New(cfg)
		p := aws_eks.NewListClustersPaginator(c.ListClustersRequest(&aws_eks.ListClustersInput{}))
		r := make([]Resource, 0)
		for p.Next(ctx) {
			page := p.CurrentPage()
			for _, cluster := range page.Clusters {
				r = append(r, Resource{
					Platform:   PlatformAWS,
					Identifier: cluster,
					Location:   cfg.Region,
					Owner:      "",
					Type:       "EKS::cluster",
				})
			}
		}
		return r, p.Err()
	},
}

// refer to https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html#Concepts.RegionsAndAvailabilityZones.Availability
// for available regions
var awsRegionPrefixes = []string{
	"us-",
	"eu-",
}

func collectAWSResources(ctx context.Context, verbose bool) ([]Resource, error) {
	log := log.New(os.Stdout, "aws: ", log.LstdFlags|log.Lmsgprefix)
	if verbose {
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
		shouldCheckRegion := false
		for _, prefix := range awsRegionPrefixes {
			if strings.HasPrefix(region.ID(), prefix) {
				shouldCheckRegion = true
				break
			}
		}
		if !shouldCheckRegion {
			continue // skip this region
		}
		if verbose {
			log.Printf("querying region %s", cfg.Region)
		}

		cfg.Region = region.ID()
		// query configured resource in region
		for rid, fetch := range awsResources {
			r, err := fetch(ctx, cfg.Copy())
			if err != nil {
				if verbose {
					log.Printf("resource fetch for '%s' failed in region %s: %v", rid, cfg.Region, err)
				}
				continue
			}
			resources = append(resources, r...)
		}
	}

	return resources, nil
}
