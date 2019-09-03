# Grafana image 

Vanilla Grafana image with one addition: embedded Sourcegraph provisioning.

# Image API

```shell script
docker run  \
    -v ${GRAFANA_DISK}:/var/lib/grafana \
    sourcegraph/grafana
```

Image expects one volume mounted:

* at `/var/lib/grafana` a data directory where logs, the Grafana db and other Grafana data files will live

Additional behavior can be controlled with 
[environmental variables](https://grafana.com/docs/installation/configuration/). 

