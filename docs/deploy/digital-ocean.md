+++
title = "Installing on Digital Ocean"
navtitle = "on Digital Ocean"
+++

To set up Sourcegraph on a new [Digital Ocean](https://www.digitalocean.com/) cloud VM, follow these steps.

* Open the [**Create Droplet**](https://cloud.digitalocean.com/droplets/new) screen on the [Digital Ocean Control Panel](https://cloud.digitalocean.com/).
* **Droplet Hostname:** Choose a valid name.
* **Select Size:** Select a droplet size with at least 2 GB of RAM.
* **Select Image:** Select the Ubuntu 14.04 x64 image.
* **Available Settings:** Check the **User Data** box and paste in the following:

{{< userdata SRC_DIGITAL_OCEAN >}}

* Click **Create Droplet**.
* In 3-4 minutes, your Sourcegraph server should be available at `http://<ip-address>`
(the instance will show you its IP address).

# Next steps

* [Configure DNS]({{< relref "config/appurl-dns.md" >}})

# Troubleshooting

* [cloud-init troubleshooting]({{< relref "troubleshooting/cloud-init.md" >}})
