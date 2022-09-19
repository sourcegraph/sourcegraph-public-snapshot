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
- Security group with default rules: `HTTP(port 80), HTTPS(port 443), and SSH(port 22)`
- The latest version of Sourcegraph

You only need to choose your VPC and SSH Key-Pair to get started.

---

## Steps

1. Choose an AWS Region in the launcher below
2. Click on the `Launch Stack` button
3. Select or create a new Key Pair for connecting to your instance securely

## Launcher
<span class="badge badge-warning">Coming soon</span> Set up a Sourcegraph instance in one click
<!-- ref: https://aws.amazon.com/blogs/devops/construct-your-own-launch-stack-url/ -->
<form class="launcher" name="launcher" action="" target="_blank">
  <select name="region" disabled>
    <option value=us-east-1#/stacks/new?">us-east-1 (N. Virginia)</option>
    <option value="us-east-2#/stacks/new?">us-east-2 (Ohio)</option>
    <option value="us-west-1#/stacks/new?">us-west-1 (N. California)</option>
    <option value="us-west-2#/stacks/new?">us-west-2 (Oregon)</option>
    <option value="ap-south-1#/stacks/new?">ap-south-1 (Asia Pacific - Mumbai)</option>
    <option value="eu-west-1#/stacks/new?">eu-west-1 (Europe - Ireland)</option>
    <option value="eu-west-2#/stacks/new?">eu-west-2 (Europe - Frankfurt)</option>
  </select>
  <input class="submit-btn" formaction="https://console.aws.amazon.com/cloudformation/home" type="image" alt="aws-oneclick-button" src="https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png" disabled/>
</form>

> NOTE: Once the instance has been created, Sourcegraph will be running at your server’s IP address, which allows you to navigate to your newly created Sourcegraph instance in your browser. You can also find the URL for your Sourcegraph instance in the 'Outputs' section of the Stack.
