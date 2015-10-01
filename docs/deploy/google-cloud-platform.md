+++
title = "Installing on Google Cloud Platform"
navtitle = "on Google Cloud Platform"
+++

To set up Sourcegraph on a new [Google Compute Engine](https://cloud.google.com/compute/) instance, follow these steps.

* Open the Google Developers Console for your project.
* In the left menu, choose **Compute** > **Compute Engine** > **VM instances** and click **New instance**.
* **Machine type:** Anything with at least 2 GB of RAM
* **Boot disk:** Ubuntu 14.04 LTS
* **Firewall:** Check the boxes for **Allow HTTP traffic** and **Allow HTTPS traffic**
* Expand the **Management, disk, networking, access & security groups** panel and set the following **Startup script**:
```
#!/bin/bash

export SRC_GOOGLE_COMPUTE_ENGINE=1

export SRC_LANGUAGE_GO=1
export SRC_LANGUAGE_JAVA=1

wget -O - https://sourcegraph.com/.download/install.sh | bash
```
* Click **Create**
* Once your instance is running, allow external access to the API on port 3100 so you can
use the `src` CLI tool to manage your Sourcegraph server. To do this, first
[install the `gcloud` CLI](https://cloud.google.com/sdk/#Quick_Start) then run ([substituting your `PROJECT-ID`](https://cloud.google.com/compute/docs/projects#projectids)):
```
gcloud --project PROJECT-ID compute firewall-rules create sourcegraph-grpc-allow-3100 --allow tcp:3100
```
* In 5 minutes, your Sourcegraph server should be available via HTTP (not HTTPS) at the VM's external IP. ***Note:** The link from the Web Console is `https://EXTERNAL-IP`, which will not work because there is no HTTPS listener. Make sure you go to `http://EXTERNAL-IP`!*

# Next steps

* [Getting started with Sourcegraph for your team]({{< relref "config/appurl-dns.md" >}})

# Troubleshooting

* [cloud-init troubleshooting]({{< relref "troubleshooting/cloud-init.md" >}})

