---
title: AWS one-click
---

<style>
.launcher {
	margin:  0.5em;
  width: 100%;
}
.launcher > select {
  width: 70%;
  font-size: 1em;
	padding: 0.2em 1em;
	margin-right: 0.25em;
  display: inline-block;
  float: left;
}
</style>

# AWS One-Click Installation for Sourcegraph

Launch a verified and pre-configured Sourcegraph instance with the following:

- Root EBS volume with 50GB of storage
- Additional EBS volume with 500GB of storage, for storing code and search indices
- AWS Security Group
- The latest version of Sourcegraph

---

## Instance Size Chart

Determine the instance type required to support the number of users and repositories you have using this table. If you fall between two sizes, choose the larger of the two.

For example, if you have 8,000 users with 80,000 repositories, your instance size would be **L**. If you have 1,000 users with 80,000 repositories, you should go with size **M**.

|                      | **XS**      | **S**       | **M**       | **L**        | **XL**       |
|----------------------|-------------|-------------|-------------|--------------|--------------|
| **Users**            | _<=_ 500    | _<=_ 1,000  | _<=_ 5,000  | _<=_ 10,000  | _<=_ 20,000  |
| **Repositories**     | _<=_ 5,000  | _<=_ 10,000 | _<=_ 50,000 | _<=_ 100,000 | _<=_ 250,000 |
| **Recommended Type** | m6a.2xlarge | m6a.4xlarge | m6a.8xlarge | m6a.12xlarge | m6a.24xlarge |
| **Minimum Type**     | m6a.2xlarge | m6a.2xlarge | m6a.4xlarge | m6a.8xlarge  | m6a.12xlarge |

Click [here](https://github.com/sourcegraph/deploy#amazon-ec2-amis) to see the completed list of AMI IDs published in each region.

## Deploy Sourcegraph

### Prerequisites

1. Create an [EC2 Key Pair](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/create-key-pairs.html)

>NOTE: The instance will launch in the default configured VPC. If your AWS user does not have a default VPC, or the `Auto-assign public IPv4 address` option is not enabled for a subnet within that VPC, please see the [Manual AMI](aws-ami.md) instructions

### Steps

1. Choose an AWS Region in the launcher below
2. Click on the `Launch Stack` button
3. Select an `SSH Keypair`
4. Select a `Sourcegraph Instance Size` according to [the sizing chart](#instance-size-chart).

 🎉 You can now start a Sourcegraph instance by clicking on the `Create Stack` button 🎉

### Launcher
<!-- ref: https://aws.amazon.com/blogs/devops/construct-your-own-launch-stack-url/ -->
<form class="launcher" name="launcher" action="" target="_blank" >
  <select name="region">
    <option value="us-east-2">us-east-2 (US East (Ohio))</option>
    <option value="us-east-1">us-east-1 (US East (N. Virginia))</option>
    <option value="us-west-1">us-west-1 (US West (N. California))</option>
    <option value="us-west-2">us-west-2 (US West (Oregon))</option>
    <option value="af-south-1">af-south-1 (Africa (Cape Town))</option>
    <option value="ap-east-1">ap-east-1 (Asia Pacific (Hong Kong))</option>
    <option value="ap-southeast-3">ap-southeast-3 (Asia Pacific (Jakarta))</option>
    <option value="ap-south-1">ap-south-1 (Asia Pacific (Mumbai))</option>
    <option value="ap-northeast-2">ap-northeast-2 (Asia Pacific (Seoul))</option>
    <option value="ap-southeast-1">ap-southeast-1 (Asia Pacific (Singapore))</option>
    <option value="ap-southeast-2">ap-southeast-2 (Asia Pacific (Sydney))</option>
    <option value="ap-northeast-1">ap-northeast-1 (Asia Pacific (Tokyo))</option>
    <option value="ca-central-1">ca-central-1 (Canada (Central))</option>
    <option value="eu-central-1">eu-central-1 (Europe (Frankfurt)</option>
    <option value="eu-west-1">eu-west-1 (Europe (Ireland))</option>
    <option value="eu-west-2">eu-west-2 (Europe (London)</option>
    <option value="eu-south-1">eu-south-1 (Europe (Milan))</option>
    <option value="eu-west-3">eu-west-3 (Europe (Paris))</option>
    <option value="eu-north-1">eu-north-1 (Europe (Stockholm))</option>
    <option value="me-south-1">me-south-1 (Middle East (Bahrain))</option>
    <option value="me-central-1">me-central-1 (Middle East (UAE))</option>
    <option value="sa-east-1">sa-east-1 (South America (São Paulo))</option>
  </select>
  <input class="submit-btn" formaction="https://console.aws.amazon.com/cloudformation/home#/stacks/quickcreate?stackName=Sourcegraph&templateURL=https://sourcegraph-cloudformation.s3.us-west-2.amazonaws.com/sg-basic.yaml" type="image" alt="aws-oneclick-button" src="https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png"/>
</form>

> NOTE: Once the instance has been created, Sourcegraph will be running at your server’s IP address, which allows you to navigate to your newly created Sourcegraph instance in your browser. You can also find the URL for your Sourcegraph instance in the 'Outputs' section of the Stack. Sourcegraph may take a few minutes initialize on first launch!

### Networking

Follow the [manual ami networking](aws-ami.md#networking) instructions to configure a domain and SSL.

## Upgrade

Follow the [manual ami](aws-ami.md#upgrade) instructions to upgrade your instance.

## Manual deploy on AWS EC2

Click [here](aws-ami.md) to view install instructions for deploying on AWS EC2 manually.