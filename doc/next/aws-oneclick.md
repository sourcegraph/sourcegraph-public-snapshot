---
title: AWS one-click
---

<style>
.launcher {
	margin: 0 0 0.5em 0;
}

.launcher select {
	font-size: 1em;
	padding: 0.2em 0 0.2em 0;
	margin-right: 0.25em;
}

.launcher .submit {
	display: inline-block;
	margin: 0.25em 0 0.25em 0;
	vertical-align: middle;
}
</style>

# AWS one-click installation

<ol>
<li>
	In one-click, the launch stacks below will create an instance with Sourcegraph pre-installed, security group with ports 22 and 80, and a volume with 200GB of storage. You will only need to choose your VPC and SSH key-pair to get started.
	<form class="launcher">
	<select>
		<option value="us-east-1">us-east-1 (N. Virginia)</option>
		<option value="us-east-2">us-east-2 (Ohio)</option>
		<option value="us-west-1">us-west-1 (N. California)</option>
		<option value="us-west-2">us-west-2 (Oregon)</option>
		<option value="ap-south-1">ap-south-1 (Asia Pacific - Mumbai)</option>
		<option value="eu-west-1">eu-west-1 (Europe - Ireland)</option>
		<option value="eu-west-2">eu-west-2 (Europe - Frankfurt)</option>
	</select>
	<a href="#TODO" class="submit">
		<img src="https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png" />
	</a>
	</form>
</li>
<li>
	Once your stack finishes creating, Retool is running at your server's IP address, which you can navigate to in your browser. Navigate to the 'Outputs' section of the Stack and you'll find the URL for your Sourcegraph instance.
</li>
</ol>
