package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_ep "github.com/aws/aws-sdk-go-v2/aws/endpoints"
	aws_ext "github.com/aws/aws-sdk-go-v2/aws/external"
	aws_ec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	aws_eks "github.com/aws/aws-sdk-go-v2/service/eks"
)

type AWSResourceFetchFunc func(context.Context, aws.Config, time.Time) ([]Resource, error)

// refer to https://docs.aws.amazon.com/sdk-for-go/v2/api/ for how to query resources under /service
var awsResources = map[string]AWSResourceFetchFunc{
	// fetch ec2 instances
	"EC2::Instances": func(ctx context.Context, cfg aws.Config, since time.Time) ([]Resource, error) {
		client := aws_ec2.New(cfg)
		pager := aws_ec2.NewDescribeInstancesPaginator(client.DescribeInstancesRequest(&aws_ec2.DescribeInstancesInput{}))
		var rs []Resource
		for pager.Next(ctx) {
			page := pager.CurrentPage()
			for _, reservation := range page.Reservations {
				for _, instance := range reservation.Instances {
					if instance.LaunchTime.After(since) {
						rs = append(rs, Resource{
							Platform:   PlatformAWS,
							Identifier: *instance.InstanceId,
							Location:   *instance.Placement.AvailabilityZone,
							Owner:      *reservation.OwnerId,
							Type:       fmt.Sprintf("EC2::Instances::%s", string(instance.InstanceType)),
							Meta: map[string]interface{}{
								"tags": instance.Tags,
							},
						})
					}
				}
			}
		}
		return rs, pager.Err()
	},
	// fetch ec2 volumes
	"EC2::Volumes": func(ctx context.Context, cfg aws.Config, since time.Time) ([]Resource, error) {
		client := aws_ec2.New(cfg)
		pager := aws_ec2.NewDescribeVolumesPaginator(client.DescribeVolumesRequest(&aws_ec2.DescribeVolumesInput{}))
		var rs []Resource
		for pager.Next(ctx) {
			page := pager.CurrentPage()
			for _, volume := range page.Volumes {
				if volume.CreateTime.After(since) {
					rs = append(rs, Resource{
						Platform:   PlatformAWS,
						Identifier: *volume.VolumeId,
						Location:   *volume.AvailabilityZone,
						Owner:      "-",
						Type:       fmt.Sprintf("EC2::Volumes::%s", string(volume.VolumeType)),
						Meta: map[string]interface{}{
							"tags":        volume.Tags,
							"attachments": volume.Attachments,
						},
					})
				}
			}
		}
		return rs, pager.Err()
	},
	// fetch kubernetes clusters
	"EKS::Clusters": func(ctx context.Context, cfg aws.Config, since time.Time) ([]Resource, error) {
		client := aws_eks.New(cfg)
		pager := aws_eks.NewListClustersPaginator(client.ListClustersRequest(&aws_eks.ListClustersInput{}))
		var rs []Resource
		for pager.Next(ctx) {
			page := pager.CurrentPage()
			for _, clusterName := range page.Clusters {
				cluster, err := client.DescribeClusterRequest(&aws_eks.DescribeClusterInput{
					Name: aws.String(clusterName),
				}).Send(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch details for cluster '%s': %w", clusterName, err)
				}
				if cluster.Cluster.CreatedAt.After(since) {
					rs = append(rs, Resource{
						Platform:   PlatformAWS,
						Identifier: *cluster.Cluster.Arn,
						Location:   cfg.Region,
						Owner:      "-",
						Type:       "EKS::Cluster",
						Meta: map[string]interface{}{
							"tags": cluster.Cluster.Tags,
						},
					})
				}
			}
		}
		return rs, pager.Err()
	},
}

// refer to https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html#Concepts.RegionsAndAvailabilityZones.Availability
// for available regions
var awsRegionPrefixes = []string{
	"us-",
	"eu-",
}

func collectAWSResources(ctx context.Context, since time.Time, verbose bool) ([]Resource, error) {
	logger := log.New(os.Stdout, "aws: ", log.LstdFlags|log.Lmsgprefix)
	if verbose {
		logger.Printf("collecting resources")
	}

	cfg, err := aws_ext.LoadDefaultAWSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to init client: %w", err)
	}

	var resources []Resource

	// query default aws regions - the only partition we are interested in is the
	// normal `aws` regions, which will always exist (other partitions include China,
	// US government, etc.)
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
			logger.Printf("querying region %s", cfg.Region)
		}

		cfg.Region = region.ID()
		// query configured resource in region
		for resourceID, fetchResource := range awsResources {
			rs, err := fetchResource(ctx, cfg.Copy(), since)
			if err != nil {
				if verbose {
					logger.Printf("resource fetch for '%s' failed in region %s: %v", resourceID, cfg.Region, err)
				}
				continue
			}
			resources = append(resources, rs...)
		}
	}

	return resources, nil
}
