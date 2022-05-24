# Sourcegraph cAdvisor

Learn more about the Sourcegraph cAdvisor distribution in the [cAdvisor documentation](https://docs.sourcegraph.com/dev/background-information/observability/cadvisor).

## Updating the image

This image **is not** currently built by CI and errors that can break builds may go undetected. This is already under the attention of the dev-experience team.

The base images for cadvisor are hosted in gcr.io/cadvisor/cadvisor. Note that the images are not tagged by architecture. That needs to be verified manually:

1. Exec in the container with `docker run -it --entrypoint /bin/sh <image name>`
2. Run `uname -a` in the container
3. Ensure it's `x86_64` architecture.

For example:

```
docker run -it --entrypoint /bin/sh gcr.io/cadvisor/cadvisor@sha256:8938726fe00fd7a3889f7c4fb50a54b728f1d02fb5f6cbdbea604824ad11ff3f
Unable to find image 'gcr.io/cadvisor/cadvisor@sha256:8938726fe00fd7a3889f7c4fb50a54b728f1d02fb5f6cbdbea604824ad11ff3f' locally
gcr.io/cadvisor/cadvisor@sha256:8938726fe00fd7a3889f7c4fb50a54b728f1d02fb5f6cbdbea604824ad11ff3f: Pulling from cadvisor/cadvisor
df9b9388f04a: Already exists
32357bb75a40: Already exists
4f4fb700ef54: Already exists
a80659b3f11d: Already exists
769b92fe592c: Already exists
6ab71a81e6dc: Already exists
Digest: sha256:8938726fe00fd7a3889f7c4fb50a54b728f1d02fb5f6cbdbea604824ad11ff3f
Status: Downloaded newer image for gcr.io/cadvisor/cadvisor@sha256:8938726fe00fd7a3889f7c4fb50a54b728f1d02fb5f6cbdbea604824ad11ff3f
/ # uname -a
Linux c74b74199c86 5.10.104-linuxkit #1 SMP Thu Mar 17 17:08:06 UTC 2022 x86_64 Linux
```
