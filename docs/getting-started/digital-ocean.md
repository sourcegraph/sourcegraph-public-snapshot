+++
title = "Installing on Digital Ocean"
linktitle = "on Digital Ocean"
+++

To set up Sourcegraph on a new [Digital Ocean](https://www.digitalocean.com/) cloud VM, follow these steps.

* Open the [**Create Droplet**](https://cloud.digitalocean.com/droplets/new) screen on the [Digital Ocean Control Panel](https://cloud.digitalocean.com/):
	* **Image:** Select the Ubuntu 14.04 x64 image.
	* **Size:** Select a droplet size with at least 4 GB of RAM (recommended).
	* **Hostname:** Choose a valid name.
	* **Additional Options:** Check the **User Data** box and paste in the following:

		```
		{{% userdata SRC_DIGITAL_OCEAN %}}
		```

* Click **Create** button.
* In 3-4 minutes, your Sourcegraph server should be available at `http://<ip-address>`
(the instance will show you its IP address).

## Questions?

* [cloud-init troubleshooting]({{< relref "troubleshooting/cloud-init.md" >}})

{{< ads_conversion >}}
