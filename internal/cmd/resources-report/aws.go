package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_ep "github.com/aws/aws-sdk-go-v2/aws/endpoints"
	aws_ext "github.com/aws/aws-sdk-go-v2/aws/external"
	aws_ec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	aws_eks "github.com/aws/aws-sdk-go-v2/service/eks"
)

type AWSResourceFetchFunc func(context.Context, aws.Config) ([]Resource, error)

// refer to https://docs.aws.amazon.com/sdk-for-go/v2/api/ for how to query resources under /service
var awsResources = []AWSResourceFetchFunc{
	// fetch ec2 instances
	func(ctx context.Context, cfg aws.Config) ([]Resource, error) {
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
	func(ctx context.Context, cfg aws.Config) ([]Resource, error) {
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
		// query configured resource in region
		for _, fetch := range awsResources {
			r, err := fetch(ctx, cfg.Copy())
			if err != nil {
				return nil, fmt.Errorf("resource fetch failed: %w", err)
			}
			resources = append(resources, r...)
		}
	}

	return resources, nil
}
