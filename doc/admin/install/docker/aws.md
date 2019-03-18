# Install Sourcegraph with Docker on AWS

<style>
div.alert-info {
    background-color: rgb(221, 241, 255);
    border-radius: 0.5em;
    padding: 0.25em 1em 0.25em 1em;
}
</style>

This tutorial shows you how to deploy Sourcegraph to a single node running on AWS.

If you're just starting out, we recommend [installing Sourcegraph locally](index.md). It takes only a few minutes and lets you try out all of the features. If you need scalability and high-availability beyond what a single-server deployment can offer, use the [Lubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph).

---

## Deploy to EC2

### Option A: use the AWS wizard

- Click **Launch Instance** from your [EC2 dashboard](https://console.aws.amazon.com/ec2/v2/home).
- Select any Amazon Linux (the first option, `ami-824c4ee2`, works fine).
- Select an appropriate instance size (we recommend t2.medium/large, depending on team size and number of repositories/languages enabled), then **Next: Configure Instance Details**
- Add the following user data (as text) in the **Advanced Details** section:

  ```
  #cloud-config
  repo_update: true
  repo_upgrade: all

  packages:
  - docker

  runcmd:
  - mkdir -p /home/ec2-user/.sourcegraph/config
  - mkdir -p /home/ec2-user/.sourcegraph/data
  - sed -i -e 's/1024/10240/g' /etc/sysconfig/docker
  - sed -i -e 's/4096/40960/g' /etc/sysconfig/docker
  - service docker start
  - usermod -a -G docker ec2-user
  - [ sh, -c, 'docker run -d --publish 80:7080 --publish 443:7443 --restart unless-stopped --volume /home/ec2-user/.sourcegraph/config:/etc/sourcegraph --volume /home/ec2-user/.sourcegraph/data:/var/opt/sourcegraph --volume /var/run/docker.sock:/var/run/docker.sock sourcegraph/server:2.12.2' ]
  ```

- Select **Next: ...** until you get to the **Configure Security Group** page, then add the default **HTTP** rule (port range "80", source "0.0.0.0/0, ::/0")
- Launch your instance, then navigate to the its public URL.
- If you have configured a DNS entry for the IP, configure `appURL` to reflect that. If `appURL` has the HTTPS protocol then Sourcegraph will get a certificate via [Let's Encrypt](https://letsencrypt.org/). For more information or alternative methods view our documentation on [TLS](../../tls_ssl.md).

### Option B: use the CLI

Use the [`aws` CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-welcome.html) to boot an EC2 instance running Sourcegraph.

First, create a `cloud-init.txt` file with user data contents as shown above or below. Then run:

`aws ec2 run-instances --image-id ami-824c4ee2 --count 1 --instance-type t2.small --key-name id_rsa --security-groups default --user-data file://cloud-init.txt`

Substitute the path to your `cloud-init.txt` file, the name of your key pair, and an appropriate security group. To start you probably want a security group which exposes port 80 and 22 (for SSH) to the public internet.

---

## Update your Sourcegraph version

To update to the most recent version of Sourcegraph (X.Y.Z), SSH into your instance and run the following:

```
docker ps # get the $CONTAINER_ID of the running sourcegraph/server container
docker rm -f $CONTAINER_ID
docker run -d ... sourcegraph/server:X.Y.Z
```

---

## Sourcegraph instances created before July 30, 2018

**The below sections only pertain to Sourcegraph instances created using this tutorial before July 30, 2018**.

This applies to you if you see the following warning on the **Admin > Code intelligence** page:

> Language server management capabilities disabled because /var/run/docker.sock was not found.

### Option A: Continue using manual code intelligence installation

Just as before July 30, 2018, you can continue manually managing code intelligence for your Sourcegraph instance if you prefer. The instructions for this have [moved here](../../../extensions/language_servers/install/aws.md).

### Option B (recommended): Upgrade to the new automatic code intelligence

Instead of manually managing code intelligence, you can upgrade to the new automatic code intelligence method.

This allows Sourcegraph to automatically set up language servers for you when new repositories are added with languages we support, in addition to allowing you (the site admin) to manage (or explicitly disable) running language servers, view their health, etc. from within the application UI on the **Admin > Code intelligence** page.

To upgrade your existing instance to use automatic code intelligence, **SSH into your Sourcegraph instance** and run the following:

1.  `docker stop $SOURCEGRAPH_CONTAINER_NAME` (find the container name using `docker ps`).
2.  Start the Docker container again using the new `docker run` command provided in the updated user-data `#cloud-config` script above. i.e.:

    ```
    docker run -d --publish 80:7080 --publish 443:7443 --restart unless-stopped --volume /home/ec2-user/.sourcegraph/config:/etc/sourcegraph --volume /home/ec2-user/.sourcegraph/data:/var/opt/sourcegraph --volume /var/run/docker.sock:/var/run/docker.sock sourcegraph/server:2.12.2
    ```

These steps only need to be performed once, and they will persist across machine restarts.

After performing these steps, you will now have automatic code intelligence! To verify, go to the **Admin > Code intelligence** page and confirm that you see Enable/Disable/restart buttons next to each language server.
