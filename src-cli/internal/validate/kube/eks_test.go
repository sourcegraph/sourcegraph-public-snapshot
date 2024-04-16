package kube

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/sourcegraph/src-cli/internal/validate"
	corev1 "k8s.io/api/core/v1"
)

func TestValidateVpc(t *testing.T) {
	cases := []struct {
		name   string
		vpc    func(vpc *types.Vpc)
		result []validate.Result
	}{
		{
			name: "valid vpc",
			vpc: func(vpc *types.Vpc) {
				vpc.State = "available"
			},
			result: []validate.Result{
				{
					Status:  validate.Success,
					Message: "VPC is validated",
				},
			},
		},
		{
			name: "invalid vpc: pending",
			vpc: func(vpc *types.Vpc) {
				vpc.State = "pending"
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "vpc.State stuck in pending state",
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			vpc := testVPC()
			if tc.vpc != nil {
				tc.vpc(vpc)
			}
			result := validateVpc(vpc)

			// test should error
			if len(tc.result) > 0 {
				if result == nil {
					t.Fatal("validate should return result")
					return
				}
				if result[0].Status != tc.result[0].Status {
					t.Errorf(
						"result status\nwant: %v\n got: %v",
						tc.result[0].Status,
						result[0].Status,
					)
				}
				if result[0].Message != tc.result[0].Message {
					t.Errorf(
						"result msg\nwant: %s\n got: %s",
						tc.result[0].Message,
						result[0].Message,
					)
				}
				return
			}

			// test should not error
			if result != nil {
				t.Fatalf("ValidateService error: %v", result)
			}
		})
	}
}

func TestValidateAddons(t *testing.T) {
	cases := []struct {
		name   string
		addons func(addons *eks.ListAddonsOutput)
		result []validate.Result
	}{
		{
			name: "should pass if 'aws-ebs-csi-driver' present in addons",
			addons: func(addonsOutput *eks.ListAddonsOutput) {
				addonsOutput.Addons = append(addonsOutput.Addons, "aws-ebs-csi-driver")
			},
			result: []validate.Result{
				{
					Status:  validate.Success,
					Message: "EKS: 'aws-ebs-csi-driver' present in addons",
				},
			},
		},
		{
			name:   "should fail if 'aws-ebs-csi-driver' not present in addons",
			addons: func(addons *eks.ListAddonsOutput) {},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "EKS: no 'aws-ebs-csi-driver' present in addons",
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			addons := testAddonOutput()
			if tc.addons != nil {
				tc.addons(addons)
			}
			result := validateAddons(addons.Addons)

			// test should error
			if len(tc.result) > 0 {
				if result == nil {
					t.Fatal("validate should return result")
					return
				}
				if result[0].Status != tc.result[0].Status {
					t.Errorf(
						"result status\nwant: %v\n got: %v",
						tc.result[0].Status,
						result[0].Status,
					)
				}
				if result[0].Message != tc.result[0].Message {
					t.Errorf(
						"result msg\nwant: %s\n got: %s",
						tc.result[0].Message,
						result[0].Message,
					)
				}
				return
			}

			// test should not error
			if result != nil {
				t.Fatalf("ValidateService error: %v", result)
			}
		})
	}
}

func TestValidateServiceAccount(t *testing.T) {
	cases := []struct {
		name   string
		sa     func(sa *corev1.ServiceAccount)
		result []validate.Result
	}{
		{
			name: "should pass if 'ebs-csi-controller-sa' present on cluster",
			sa: func(sa *corev1.ServiceAccount) {
				sa.ObjectMeta.Name = "ebs-csi-controller-sa"
				sa.ObjectMeta.Namespace = "kube-system"
			},
			result: []validate.Result{
				{
					Status:  validate.Success,
					Message: "EKS: 'ebs-csi-controller-sa' present on cluster",
				},
			},
		},
		{
			name: "should fail if 'ebs-csi-controller-sa' not present on cluster",
			sa:   func(sa *corev1.ServiceAccount) {},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "EKS: no 'ebs-csi-controller-sa' present on cluster",
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			sa := testServiceAccount()
			if tc.sa != nil {
				tc.sa(sa)
			}
			result := validateServiceAccount(sa)

			// test should error
			if len(tc.result) > 0 {
				if result == nil {
					t.Fatal("validate should return result")
					return
				}
				if result[0].Status != tc.result[0].Status {
					t.Errorf(
						"result status\nwant: %v\n got: %v",
						tc.result[0].Status,
						result[0].Status,
					)
				}
				if result[0].Message != tc.result[0].Message {
					t.Errorf(
						"result msg\nwant: %s\n got: %s",
						tc.result[0].Message,
						result[0].Message,
					)
				}
				return
			}

			// test should not error
			if result != nil {
				t.Fatalf("ValidateService error: %v", result)
			}
		})
	}
}

func TestValidateRolePolicy(t *testing.T) {
	cases := []struct {
		name   string
		rp     func(rp *iam.ListAttachedRolePoliciesOutput)
		result []validate.Result
	}{
		{
			name: "should pass if 'AmazonEBSCSIDriverPolicy' bound to role",
			rp: func(rp *iam.ListAttachedRolePoliciesOutput) {
				AmazonEBSCSIDriverPolicy := "AmazonEBSCSIDriverPolicy"
				rp.AttachedPolicies = append(rp.AttachedPolicies, iamTypes.AttachedPolicy{
					PolicyName: &AmazonEBSCSIDriverPolicy,
				})
			},
			result: []validate.Result{
				{
					Status:  validate.Success,
					Message: "EKS: 'AmazonEBSCSIDriverPolicy' bound to role",
				},
			},
		},
		{
			name: "should fail if 'AmazonEBSCSIDriverPolicy' not bound to role",
			rp:   func(rp *iam.ListAttachedRolePoliciesOutput) {},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "EKS: no 'AmazonEBSCSIDriverPolicy' bound to role",
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rp := testEBSCSIRole()
			if tc.rp != nil {
				tc.rp(rp)
			}

			var result []validate.Result

			if len(rp.AttachedPolicies) == 0 {
				shouldFail := "should fail"
				result = validateRolePolicy(RolePolicy{
					PolicyName: &shouldFail,
					PolicyArn:  &shouldFail,
				})
			} else {
				rolePolicy := RolePolicy{
					PolicyName: rp.AttachedPolicies[0].PolicyName,
					PolicyArn:  rp.AttachedPolicies[0].PolicyArn,
				}

				result = validateRolePolicy(rolePolicy)
			}

			// test should error
			if len(tc.result) > 0 {
				if result == nil {
					t.Fatal("validate should return result")
					return
				}
				if result[0].Status != tc.result[0].Status {
					t.Errorf(
						"result status\nwant: %v\n got: %v",
						tc.result[0].Status,
						result[0].Status,
					)
				}
				if result[0].Message != tc.result[0].Message {
					t.Errorf(
						"result msg\nwant: %s\n got: %s",
						tc.result[0].Message,
						result[0].Message,
					)
				}
				return
			}

			// test should not error
			if result != nil {
				t.Fatalf("ValidateService error: %v", result)
			}
		})
	}
}

// helpers
func testVPC() *types.Vpc {
	return &types.Vpc{
		State: "available",
	}
}

func testAddonOutput() *eks.ListAddonsOutput {
	return &eks.ListAddonsOutput{}
}

func testServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{}
}

func testEBSCSIRole() *iam.ListAttachedRolePoliciesOutput {
	return &iam.ListAttachedRolePoliciesOutput{
		AttachedPolicies: []iamTypes.AttachedPolicy{},
	}
}
