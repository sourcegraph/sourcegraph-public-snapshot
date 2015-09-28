+++
title = "Installing on Google Cloud Platform"
navtitle = "on Google Cloud Platform"
+++

To set up Sourcegraph on a new [Google Compute Engine](https://cloud.google.com/compute/) instance, follow these steps.

* Open the Google Developers Console for your project.
* In the left menu, choose **Compute** > **Compute Engine** > **VM instances** and click **New instance**.
* **Machine type:** Anything with at least 1 GB of RAM
* **Boot disk:** Ubuntu 14.04 LTS
* **Firewall:** Check the boxes for **Allow HTTP traffic** and **Allow HTTPS traffic**
* Expand the **Management, disk, networking, access & security groups** panel and set the following **Startup script**:
```
{{% userdata %}}
```
* Allow external access to the API on port 3100 so you can use the `src` CLI tool to manage your Sourcegraph server:
```
gcloud --project PROJECT-ID compute firewall-rules create sourcegraph-grpc-allow-3100 --allow tcp:3100
```
* In 5 minutes, your Sourcegraph server should be available via HTTP (not HTTPS) at the VM's external IP. ***Note:** The link from the Web Console is `https://EXTERNAL-IP`, which will not work because there is no HTTPS listener. Make sure you go to `http://EXTERNAL-IP`!*

Note: [Enabling TLS]({{< relref "config/https.md" >}}) is *strongly
recommended*, to avoid leaking credentials.

# Next steps

* [Getting started with Sourcegraph for your team]({{< relref "getting-started/index.md" >}})

# Troubleshooting

* [cloud-init troubleshooting]({{< relref "troubleshooting/cloud-init.md" >}})

