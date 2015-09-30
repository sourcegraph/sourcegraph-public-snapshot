+++
title = "Installing on Digital Ocean"
navtitle = "on Digital Ocean"
+++

To set up Sourcegraph on a new [Digital Ocean](https://www.digitalocean.com/) cloud VM, follow these steps.

* Open the [**Create Droplet**](https://cloud.digitalocean.com/droplets/new) screen on the [Digital Ocean Control Panel](https://cloud.digitalocean.com/).
* **Droplet Hostname:** Choose a valid, externally accessible DNS name for your Sourcegraph server. For example, if you choose `src.mycompany.com`, then you must also set up a DNS record (or a temporary `/etc/hosts` entry) so that `src.mycompany.com` points to your droplet's IP (which you can view shortly after the droplet is created).
* **Select Size:** Select a droplet size with at least 2 GB of RAM.
* **Select Image:** Select the Ubuntu 14.04 x64 image.
* **Available Settings:** Check the **User Data** box and paste in the following:
```
{{% userdata %}}
```
* Click **Create Droplet**.
* After creating the droplet, be sure to set up a DNS record (or `/etc/hosts` entry) for your Droplet's hostname.
* In 3-4 minutes, your Sourcegraph server should be available at `http://HOSTNAME`.

# Next steps

* [Getting started with Sourcegraph for your team]({{< relref "config/appurl-dns.md" >}})

# Troubleshooting

You can access your server via SSH (as the `root` user) using the SSH
keypair you chose, or the root password that Digital Ocean emailed
you.

* [cloud-init troubleshooting]({{< relref "troubleshooting/cloud-init.md" >}})
