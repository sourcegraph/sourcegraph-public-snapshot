# Sourcegraph AWS AMI Instances

[Amazon Machine Images (AMIs)](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instances-and-amis.html) allows you to start a verified and pre-configured Sourcegraph instance with everything you need for your instance size in just a few clicks.

A Sourcegraph AWS AMI instance includes:

- Root EBS volume with 50GB of storage
- Additional EBS volume with 500GB of storage for storing code and search indices
  - Storage space is configurable and expandable
- A specific version of Sourcegraph based on the selected AMI
- Resource requirements are configured according to your selected instance size

You only need to choose your VPC and SSH Key-Pair to get started.

See the [official docs](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instances-and-amis.html) to learn more about instances and AMIs.

---

## Instance size

You should select an AMI that works with the amount of users and repositories you have. 

For example, if you have 8,000 users with 80,000 repositories, your instance size would be **M**. 

If you have 1,000 users with 80,000 users, you should still go with size **M**.

| **Size**           | **XS**       | **S**       | **M**       | **L**        | **XL**       |
|:------------------:|:-----------:|:-----------:|:-----------:|:------------:|:------------:|
| **Users**          | 500         | 1,000       | 5,000       | 10,000       | 20,000       |
| **Repositories**   | 5,000       | 10,000      | 50,000      | 100,000      | 250,000      |
| **AMI Name**       | sourcegraph-XS (__v4.0.0__) m6a.2xlarge | sourcegraph-S (__v4.0.0__) m6a.4xlarge | sourcegraph-M (__v4.0.0__) m6a.8xlarge | sourcegraph-L (__v4.0.0__) m6a.12xlarge | sourcegraph-XL (__v4.0.0__) m6a.24xlarge |

<span class="badge badge-critical">IMPORTANT</span> Replace __4.0.0__ with the version number of your choice. **Versions below v4.0.0 are not supported.**

### AMI names

For example, below are the names of the AMIs for version 4.0.0:

- XS: [sourcegraph-XS (v4.0.0) m6a.2xlarge](https://console.aws.amazon.com/ec2#ImageDetails:imageId=ami-0ee5cdc5e89a4bee2)
- S : sourcegraph-S (v4.0.0) m6a.4xlarge
- M : sourcegraph-M (v4.0.0) m6a.8xlarge
- L: [sourcegraph-L (v4.0.0) m6a.12xlarge](https://console.aws.amazon.com/ec2#ImageDetails:imageId=ami-021db30b6db9b0634)
- XL: [sourcegraph-XL (v4.0.0) m6a.24xlarge](https://console.aws.amazon.com/ec2#ImageDetails:imageId=ami-04b10e0fabedb6eac)

## Instance types

Here is a list of suggestions of instance types for each instance size.

The recommended instance type has better performance due to allocated resources.

Please select the instance type according to the table below for your instance size, and do not go below the suggested minimum instance type.

| **Size**           | **S**       | **M**       | **L**        | **XL**       | **2XL**      |
|:------------------:|:-----------:|:-----------:|:------------:|:------------:|:------------:|
| **Users**          | 1,000       | 5,000       | 10,000       | 20,000       | 40,000       |
| **Repositories**   | 10,000      | 50,000      | 100,000      | 250,000      | 500,000      |
| **Recommended**    | m6a.4xlarge | m6a.8xlarge | m6a.12xlarge | m6a.24xlarge | m6a.48xlarge |
| **Minimum**        | m6a.2xlarge | m6a.4xlarge | m6a.8xlarge  | m6a.12xlarge | m6a.48xlarge |

> NOTE: You can resize your instance anytime by [changing the instance type](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-resize.html) associated with your instance size. If you need to change beyond the minimum or maximum supports of your current instance type, you can follow our [upgrade steps](#upgrade) in this page to start a new instance using the new instance size with its associated instance typ.  Please make sure the volumes are backed up before switching instance type.

---

## Deploy

![image](https://user-images.githubusercontent.com/68532117/192365588-05c4d828-9d15-4c2d-a2e7-169b5f3ce20f.png)

1. Open the [Amazon EC2 AMIs console](https://console.aws.amazon.com/ec2/home#Images:visibility=public-images;owner=185007729374;Name=production;)
2. Choose **Public images** from the dropdown menu next to the search bar
3. Enter `Owner = 185007729374` and `Name=production` or the [name of the AMI](#ami-names) in the search bar
4. Select the AMI published by Sourcegraph for your instance size
5. Click **Launch instance from AMI**

Alternatively, you can search for the AMIs using  [name of the AMI](#ami-names) in the community AMIs page when launching an instance:
![image](https://user-images.githubusercontent.com/68532117/192366274-5e75eaae-1f45-4f12-bae9-13bfea0a2cb2.png)


Once you've been redirected to the `Launch an instance` page...

1. Name your instance
2. Select an **instance type**
3. **Key pair (login)**: Select or create a new Key Pair for connecting to your instance securely
4. **Network settings**: 
   - Select a **Security Group** for the instance, or create one with the following rules:
     - Default HTTP rule: port range 80, source 0.0.0.0/0, ::/0
     - Default HTTPS rule: port range 443, source 0.0.0.0/0, ::/0
     - Default SSH rule: port range 22, source 0.0.0.0/0, ::/0
5. **Configure storage**:
   - Root Volume: 50GB/gp2
   - EBS Volume: 500GB/gp3 (20% more than the total size of all your repositories)
6. Click **Launch instance** to launch

Your Sourcegraph instance should be ready in the next few minutes. 

---

## Upgrade

> WARNING: This upgrade process works with **Sourcegraph AWS AMI instances only**

<span class="badge badge-critical">IMPORTANT</span> **Back up your volumes before each upgrade**

Please take time to review the following before proceeding with the upgrades:

- [Changelog](https://docs.sourcegraph.com/CHANGELOG)
- [Update policy](https://docs.sourcegraph.com/admin/updates#update-policy)
- [Update notes](https://docs.sourcegraph.com/admin/updates/kubernetes)
- [Multi-version upgrade procedure](https://docs.sourcegraph.com/admin/updates/kubernetes#multi-version-upgrade-procedure)

#### Step 1: Stop the current instance

1. Stop your current Sourcegraph AMI instance
   - Go to the ECS console for your instance
   - Click Instance State to Stop Instance
2. Detach the non-root data volume (Device name: /dev/sdb/)
   - Go to the Storage section in your instance console
   - Find the volume with the device name **/dev/sdb**
   - Select the volume, then click Actions to **Detach Volume**
   - Give the volume a name for identification purposes
3. Make a note of the VPC name

#### Step 2: Launch a new instance

1. Launch a new Sourcegraph instance from an AMI with the latest version of Sourcegraph
2. Name the instance
3. Select the appropriate **instance type**
4. Under **Key Pair**
  - Select the **Key Pair** used by the old instance
5. Under **Network settings**
   - Select the **Security Group** used by the old instance
6. Under **Configure storage**
 - Remove the **second** EBS volume
7. After reviewing the settings, click **Launch Instance**
8. Attach the detached volume to the new instance
   - Go to the Volumes section in your ECS Console
   - Select the volume you've detached earlier
   - Click **Actions > Attach Volume**
9. On the Attach volume page:
  - **Instance**: select the new Sourcegraph AMI instance
  - **Device name**: /dev/sdb

10\. Reboot the new instance

You can terminate the stopped Sourcegraph AMI instance once you have confirmed the new instance is up and running.

### Downgrade

Please refer to the upgrade steps above for downgrading your instance. 

---

## Network

If you have access to your instance, you can follow our [HTTP and HTTPS/SSL configuration guide](../../../admin/http_https_configuration.md#sourcegraph-via-docker-compose-caddy-2) to set up HTTP and HTTPS connections.

### AWS Load Balancer

#### Set-up Overview:

> NOTE: You must own a DNS before you can proceed with the following steps.

1. Request a certificate for the DNS via AWS Certificate Manager
2. Create a target group for HTTPS Port 443 that links to the instance Port 443
3. Create a new subnet inside the instance VPC
4. Create a new Load Balancer > Application Load Balancer

#### Step 1: Request certificate

- **Domain names**: Fully qualified domain name: your domain
- **Select validation method**: DNS validation - recommended
 
After the certificate has been created, you will need to attach the CNAME name and values to your DNS.

If your DNS is hosted in AWS route 53:

1. Click **Create record in route 53** in the certificate dashboard
2. Select the DNS you would like to attach the certificate to
3. Click **Create records** once you have verified the information is correct
4. Wait 30 mins before the validation is completed

#### Step 2: Create a target group

1. Click **Create a target group** on your EC2 Target groups dashboard
   - Choose a target type: Instance
   - Target group name: name
   - Protocol: HTTPS
   - Port: 443
   - VPC: Where your instance is located
   - Protocol version: HTTP2
   - Health checks: Use Default
2. Click **Include as pending below**

#### Step 3: Create subnets

Click **Create subnet** in your VPC subnets dashboard:

- **VPC ID**: Selected the VPC that the instance is in
- **Subnet name**: name the subnet
- **Availability Zone**: select an availability zone that is different from the current zone
- Click **Create subnet**

#### Step 4: Create an Application Load Balancer

- **Basic configuration**
  - Load balancer name: name
  - Scheme: Internet-facing
  - IP address type: IPv4
- **Network mapping**
  - VPC: Selected the VPC that the instance is in
  - Mapping: Select two subnets associated with the selected VPC
  - Security groups
  - Security groups: Make sure only the security group associated with the instance is selected
- **Listeners and routing**
  - Protocol: HTTPS
  - Port: 443
  - Default action: Select the HTTPS target group created for the instance

---

## Storage and Backups

We strongly recommend you taking [snapshots of the entire EBS volume](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html) on an [automatic, scheduled basis](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snapshot-lifecycle.html).

---

## Manual deploy on AWS EC2

Click [here](aws.md) to view install instructions for deploying on AWS EC2 manually.

--- 

## Resources

- [Increase the size of an Amazon EBS volume on an EC2 instance](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/modify-ebs-volume-on-instance.html)
- [Change the instance type](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-resize.html)
