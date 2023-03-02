package kube

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/sourcegraph/src-cli/internal/validate"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type EbsTestObjects struct {
	addons             []string
	serviceAccount     *corev1.ServiceAccount
	serviceAccountRole string
	ebsRolePolicy      RolePolicy
}

type RolePolicy struct {
	PolicyName *string
	PolicyArn  *string
}

func GenerateAWSClients(ctx context.Context) Option {
	eksConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("error while loading config: %s\n", err)
	}

	ec2Config, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("error while loading config: %s\n", err)
	}

	iamConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("error while loading config: %s\n", err)
	}

	return func(config *Config) {
		config.eks = true
		config.eksClient = eks.NewFromConfig(eksConfig)
		config.ec2Client = ec2.NewFromConfig(ec2Config)
		config.iamClient = iam.NewFromConfig(iamConfig)
	}
}

func EksVpc(ctx context.Context, config *Config) ([]validate.Result, error) {
	var results []validate.Result
	if config.ec2Client == nil {
		results = append(results, validate.Result{
			Status:  validate.Failure,
			Message: "EKS: validate VPC failed",
		})
	}

	inputs := &ec2.DescribeVpcsInput{}
	outputs, err := config.ec2Client.DescribeVpcs(ctx, inputs)

	if err != nil {
		results = append(results, validate.Result{
			Status:  validate.Failure,
			Message: "EKS: Validate VPC failed",
		})

		return results, nil
	}

	if len(outputs.Vpcs) == 0 {
		results = append(results, validate.Result{
			Status:  validate.Failure,
			Message: "EKS: Validate VPC failed: No VPC configured",
		})

		return results, nil
	}

	for _, vpc := range outputs.Vpcs {
		r := validateVpc(&vpc)
		results = append(results, r...)
	}

	return results, nil
}

func validateVpc(vpc *types.Vpc) (result []validate.Result) {
	state := vpc.State

	if state == "available" {
		result = append(result, validate.Result{
			Status:  validate.Success,
			Message: "VPC is validated",
		})

		return result
	}

	result = append(result, validate.Result{
		Status:  validate.Failure,
		Message: "vpc.State stuck in pending state",
	})

	return result
}

func EksEbsCsiDrivers(ctx context.Context, config *Config) ([]validate.Result, error) {
	var results []validate.Result
	var ebsTestParams EbsTestObjects

	addons, err := getAddons(ctx, config.eksClient)
	if err != nil {
		results = append(results, validate.Result{
			Status:  validate.Failure,
			Message: "EKS: could not validate ebs in addons",
		})

		return results, err
	}

	ebsServiceAccount, err := getEBSServiceAccount(ctx, config.clientSet)
	if err != nil {
		results = append(results, validate.Result{
			Status:  validate.Failure,
			Message: "EKS: could not validate ebs service account",
		})

		return results, err
	}

	ebsServiceAccountRole, err := getEBSServiceAccountRole(ebsServiceAccount)
	if err != nil {
		results = append(results, validate.Result{
			Status:  validate.Failure,
			Message: "EKS: could not validate ebs service account attached to role",
		})

		return results, err
	}

	ebsRolePolicy, err := getEBSCSIPolicy(ctx, config.iamClient, ebsServiceAccountRole)
	if err != nil {
		results = append(results, validate.Result{
			Status:  validate.Failure,
			Message: "EKS: could not validate ebs role policy",
		})

		return results, err
	}

	ebsTestParams.addons = addons
	ebsTestParams.serviceAccount = ebsServiceAccount
	ebsTestParams.serviceAccountRole = ebsServiceAccountRole
	ebsTestParams.ebsRolePolicy = ebsRolePolicy

	result := validateEbsCsiDrivers(ebsTestParams)
	results = append(results, result...)

	return results, nil
}

func validateEbsCsiDrivers(testers EbsTestObjects) (result []validate.Result) {
	result = append(result, validateAddons(testers.addons)...)
	result = append(result, validateServiceAccount(testers.serviceAccount)...)
	result = append(result, validateServiceAccountRole(testers.serviceAccountRole)...)
	result = append(result, validateRolePolicy(testers.ebsRolePolicy)...)

	return result
}

func validateAddons(addons []string) (result []validate.Result) {
	for _, addon := range addons {
		if addon == "aws-ebs-csi-driver" {
			result = append(result, validate.Result{
				Status:  validate.Success,
				Message: "EKS: 'aws-ebs-csi-driver' present in addons",
			})
			return result
		}
	}

	result = append(result, validate.Result{
		Status:  validate.Failure,
		Message: "EKS: no 'aws-ebs-csi-driver' present in addons",
	})

	return result
}

func validateServiceAccount(serviceAccount *corev1.ServiceAccount) (result []validate.Result) {
	if serviceAccount.Name == "ebs-csi-controller-sa" {
		result = append(result, validate.Result{
			Status:  validate.Success,
			Message: "EKS: 'ebs-csi-controller-sa' present on cluster",
		})
		return result
	}

	result = append(result, validate.Result{
		Status:  validate.Failure,
		Message: "EKS: no 'ebs-csi-controller-sa' present on cluster",
	})

	return result
}

func validateServiceAccountRole(serviceAccountRole string) (result []validate.Result) {
	if serviceAccountRole != "" {
		result = append(result, validate.Result{
			Status:  validate.Success,
			Message: "EKS: role attached to 'ebs-csi-controller-sa' service account",
		})
		return result
	}

	result = append(result, validate.Result{
		Status:  validate.Failure,
		Message: "EKS: no role attached to 'ebs-csi-controller-sa' service account",
	})

	return result
}

func validateRolePolicy(rolePolicy RolePolicy) (result []validate.Result) {
	if *rolePolicy.PolicyName == "AmazonEBSCSIDriverPolicy" {
		result = append(result, validate.Result{
			Status:  validate.Success,
			Message: "EKS: 'AmazonEBSCSIDriverPolicy' bound to role",
		})

		return result
	}

	result = append(result, validate.Result{
		Status:  validate.Failure,
		Message: "EKS: no 'AmazonEBSCSIDriverPolicy' bound to role",
	})

	return result
}

func getAddons(ctx context.Context, client *eks.Client) ([]string, error) {
	clusterName := getClusterName(ctx, client)
	inputs := &eks.ListAddonsInput{ClusterName: clusterName}
	outputs, err := client.ListAddons(ctx, inputs)

	if err != nil {
		return nil, err
	}

	return outputs.Addons, nil
}

func getEBSServiceAccount(ctx context.Context, client *kubernetes.Clientset) (*corev1.ServiceAccount, error) {
	serviceAccounts := client.CoreV1().ServiceAccounts("kube-system")
	ebsSA, err := serviceAccounts.Get(
		ctx,
		"ebs-csi-controller-sa",
		metav1.GetOptions{},
	)

	if err != nil {
		return nil, err
	}

	return ebsSA, nil
}

func getEBSServiceAccountRole(sa *corev1.ServiceAccount) (string, error) {
	annotations := sa.GetAnnotations()
	if _, ok := annotations["eks.amazonaws.com/role-arn"]; !ok {
		return "", errors.Newf(
			"%s no role attached to service account",
			validate.FailureEmoji,
		)
	}

	roleArn := strings.Split(annotations["eks.amazonaws.com/role-arn"], "/")
	if len(roleArn) != 2 {
		return "", errors.Newf(
			"%s value of 'eks.amazonaws.com/role-arn' invalid",
			validate.FailureEmoji,
		)
	}

	ebsControllerSA := roleArn[1]
	return ebsControllerSA, nil
}

// getEbsCSIPolicy
func getEBSCSIPolicy(ctx context.Context, client *iam.Client, sa string) (RolePolicy, error) {
	inputs := iam.ListAttachedRolePoliciesInput{RoleName: &sa}
	outputs, err := client.ListAttachedRolePolicies(ctx, &inputs)

	if err != nil {
		return RolePolicy{}, err
	}

	if len(outputs.AttachedPolicies) == 0 {
		return RolePolicy{}, nil
	}

	var policyName *string
	for _, policy := range outputs.AttachedPolicies {
		policyName = policy.PolicyName
		if *policyName == "AmazonEBSCSIDriverPolicy" {
			return RolePolicy{
				PolicyName: policy.PolicyName,
				PolicyArn:  policy.PolicyArn,
			}, nil
		}
	}

	return RolePolicy{}, nil
}

func getClusterName(ctx context.Context, client *eks.Client) *string {
	home := homedir.HomeDir()
	pathToKubeConfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: pathToKubeConfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()

	if err != nil {
		fmt.Printf("error while checking current context: %s\n", err)
		return nil
	}

	currentContext := strings.Split(config.CurrentContext, "/")
	clusterName := currentContext[len(currentContext)-1]

	return &clusterName
}
