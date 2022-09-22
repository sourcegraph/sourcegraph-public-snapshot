# Sourcegraph AWS AMI Instances

Launch a verified and pre-configured Sourcegraph instance with the following setup:

- Root EBS volume with 50GB of storage
- Additional EBS volume with 500GB (configurable) of storage, for storing code and search indices
- The latest version of Sourcegraph
- Resource requirements are configured according to your selected instance size

You only need to choose your VPC and SSH Key-Pair to get started.

---

## Instance sizes

| Size          | L                       | XL                      | 2XL                     |
|---------------|-------------------------|-------------------------|-------------------------|
| Users         | 10,000                  | 20,000                  | 40,000                  |
| Repositories  | 100,000                 | 250,000                 | 500,000                 |
| Instance Type | m6a.12xlarge            | m6a.24xlarge            | m6a.48xlarge            |
| AMI Name      | sourcegraph-placeholder | sourcegraph-placeholder | sourcegraph-placeholder |

---

## Deploy

![image](https://user-images.githubusercontent.com/68532117/191854109-8b81abb4-925d-436d-b14f-91607c852f7b.png)

1. Open the [Amazon EC2 AMIs console](https://console.aws.amazon.com/ec2/home#Images:visibility=public-images)
2. Choose **Public images** from the dropdown menu next to the search bar
3. Enter `Owner alias = sourcegraph` in the search bar
4. Select the AMI published by Sourcegraph for your instance size
5. Click **Launch instance from AMI**

Once you've been redirected to the `Launch an instance` page...

1. Name your instance
2. Select an **instance type**
3. **Key pair (login)**: Select or create a new Key Pair for connecting to your instance securely
4. **Network settings**: 
   - Click on the **Edit** button in the `Network setting` to enable `auto-assign public IP` 
   - Select a **Security Group** for the instance, or create one with the following rules:
     - Default HTTP rule: port range 80, source 0.0.0.0/0, ::/0
     - Default HTTPS rule: port range 443, source 0.0.0.0/0, ::/0
     - Default SSH rule: port range 22, source 0.0.0.0/0, ::/0
5. **Configure storage**:
   - Root Volume: 50GB/gp2
   - EBS Volume: 500GB/gp3(20% more than the total size of all your repositories)
6. Click **Launch instance** to launch

Your Sourcegraph instance should be ready in the next few minutes. 

You can navigate to the public IP address assigned to the EC2 node to access your newly created instance.

>NOTE: Look for the **IPv4 Public IP** value in your EC2 instance page under the *Description* panel.

---

## Upgrade

> WARNING: This upgrade process works with **Sourcegraph AWS AMI instances only**

Recommended: Please back up your volume before proceeding with the upgrades

#### Step 1: Terminate the current instance

1. Stop your current Sourcegraph AMI instance
   - Go to the ECS console for your instance
   - Click Instance State to Stop Instance
2. Detach the non-root data volume (Device name: /dev/sdb/)
   - Go to the Storage section in your instance console
   - Find the volume with the device name /dev/sdb
   - Select the volume, then click Actions to Detach Volume
   - Give the volume a name for identification purposes
3. Take a note of the VPC name
4. Terminate the stopped Sourcegraph AMI instance

#### Step 2: Launch a new instance

1. Launch a new Sourcegraph instance from an AMI with the latest version of Sourcegraph
2. Name the instance
3. Select the appropriate **instance type**
4. Under **Key Pair**
  - Select the Key Pair used by the old instance
5. Under **Network settings**
   - Click Edit > enable Auto-assign public IP
   - Select the Security Group used by the old instance
6. Under **Configure storage**
 - Remove the second EBS volume
7. After reviewing the settings, click **Launch Instance**
8. Attach the detached volume to the new instance
   - Go to the Volumes section in your ECS Console
   - Select the volume you've detached earlier
   - Click **Actions > Attach Volume**
9. On the Attach volume page:
  - **Instance**: select the new Sourcegraph AMI instance
  - **Device name**: /dev/sdb

10\. Reboot the new instance

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

## Manual deploy on AWS EC2

Click [here](aws.md) to view install instructions for deploying on AWS EC2 manually.
