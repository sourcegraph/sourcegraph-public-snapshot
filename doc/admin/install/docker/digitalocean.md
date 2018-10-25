# Install Sourcegraph with Docker on DigitalOcean

<style>
div.alert-info {
    background-color: rgb(221, 241, 255);
    border-radius: 0.5em;
    padding: 0.25em 1em 0.25em 1em;
}
</style>

This tutorial shows you how to deploy Sourcegraph to a single node running on DigitalOcean.

If you're just starting out, we recommend [installing Sourcegraph locally](index.md). It takes only a few minutes and lets you try out all of the features. If you need scalability and high-availability beyond what a single-server deployment can offer, use the [Lubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph).

---

## Use the "Create Droplets" wizard

[Open your DigitalOcean dashboard](https://cloud.digitalocean.com/droplets/new) to create a new droplet

- **Choose an image -** Select the **One-click apps** tab and then choose Docker
- **Choose a size -** We recommend at least 4GB RAM and 2 CPU, more depending on team size and number of repositories/languages enabled.
- **Select additional options -** Check "User data" and paste in the following:

  ```
  #cloud-config
  repo_update: true
  repo_upgrade: all

  runcmd:
  - mkdir -p /root/.sourcegraph/config
  - mkdir -p /root/.sourcegraph/data
  - [ sh, -c, 'docker run -d --publish 80:7080 --publish 443:7443 --restart unless-stopped --volume /root/.sourcegraph/config:/etc/sourcegraph --volume /root/.sourcegraph/data:/var/opt/sourcegraph --volume /var/run/docker.sock:/var/run/docker.sock sourcegraph/server:2.12.2' ]
  ```

- Launch your instance, then navigate to its IP address.

- If you have configured a DNS entry for the IP, configure `appURL` to reflect that. If `appURL` has the HTTPS protocol then Sourcegraph will get a certificate via [Let's Encrypt](https://letsencrypt.org/). For more information or alternative methods, see "[Using TLS/SSL](../../tls_ssl.md)".

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

Just as before July 30, 2018, you can continue manually managing code intelligence for your Sourcegraph instance if you prefer. The instructions for this have [moved here](../../../extensions/language_servers/install/digitalocean.md).

### Option B (recommended): Upgrade to the new automatic code intelligence

Instead of manually managing code intelligence, you can upgrade to the new automatic code intelligence method.

This allows Sourcegraph to automatically set up language servers for you when new repositories are added with languages we support, in addition to allowing you (the site admin) to manage (or explicitly disable) running language servers, view their health, etc. from within the application UI on the **Admin > Code intelligence** page.

To upgrade your existing instance to use automatic code intelligence, **SSH into your Sourcegraph instance** and run the following:

1.  `docker stop $SOURCEGRAPH_CONTAINER_NAME` (find the container name using `docker ps`).
2.  Start the Docker container again using the new `docker run` command provided in the updated user-data `#cloud-config` script above. i.e.:

    ```
    docker run -d --publish 80:7080 --publish 443:7443 --restart unless-stopped --volume /root/.sourcegraph/config:/etc/sourcegraph --volume /root/.sourcegraph/data:/var/opt/sourcegraph --volume /var/run/docker.sock:/var/run/docker.sock sourcegraph/server:2.12.2
    ```

These steps only need to be performed once, and they will persist across machine restarts.

After performing these steps, you will now have automatic code intelligence! To verify, go to the **Admin > Code intelligence** page and confirm that you see Enable/Disable/restart buttons next to each language server.
