# Launch Sourcegraph from an AWS AMI

<span class="badge badge-warning">Coming soon</span> Launch a verified and pre-configured Sourcegraph instance from an approved AWS AMI, created specifically for your instance size.

- A Sourcegraph AMI instance includes:
  - Latest version of Sourcegraph
  - A EBS Volume as the root disk with 50GB of storage
  - A EBS Volume as the data store with 500GB of storage

You will only need to configure a security group and SSH key-pair to get started.

### Launch

1. <span class="badge badge-note">RECOMMENDED</span> Create a [Amazon Relational Database Service (RDS)](https://aws.amazon.com/rds/)

2. Launch an instance from a Sourcegraph AMI for your instance size

3. <span class="badge badge-note">RECOMMENDED</span> Select or create a new `Key Pair` for connecting to your instance securely

4. <span class="badge badge-note">RECOMMENDED</span> Select a `Security Group` for the instance, or create a new one

5. <span class="badge badge-note">RECOMMENDED</span> Enter your RDS IP address in the **User data** text box under `Advanced`

### Upgrade

Please backup your volumes before proceeding with the upgrades.

> WARNING: This guide only works with Sourcegraph AMI instances.

#### Step 1: Terminate the current instance

1. Stop your current Sourcegraph AMI instance
  
  - Go to the ECS console for your instance
  
  - Click `Instance State` to `Stop Instance`

2. Detach the non-root data volume /dev/sdb/
   
  - Go to the `Storage` section in your instance console
  
  - Find the volume with the device name `/dev/sdb`
  
  - Select the volume, then click `Actions` to `Detach Volume`
  
  - Give the volume a name for identification purposes

3. Write down the private IP address and the VPC name of the stopped instance

4. Terminate the stopped Sourcegraph AMI instance

#### Step 2: Launch a new instance

1. Launch a new Sourcegraph instance from an AMI with latest version of Sourcegraph
 
2. Name the instance

3. Select the appropriate `instance type`

4. Under `Key Pair`
   - Select the `Key Pair` used by the old instance

5. Under `Network settings`
  - Click `Edit` > `enable Auto-assign public IP`
  
  - Select the Security Group used by the old instance

6. Under `Configure storage`
  - Remove the **second** EBS volume

7. After reviewing the settings, click `Launch Instance`

8. Attach the detached volume to the new instance
   
  - 1\. Go to the `Volumes` section in your ECS Console
  
  - 2\. Select the volume you've detached earlier
  
  - 3\. Click `Actions` > `Attach Volume`

9.  On the Attach volume page:
  
  - Instance: `select the new Sourcegraph AMI instance`
  
  - Device name: `/dev/sdb`

10. Restart the new instance
