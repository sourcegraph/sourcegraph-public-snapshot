+++
title = "Deploying on Kubernetes"
linktitle = "on Kubernetes"
+++

Sourcegraph can be deployed on any Kubernetes cluster. This guide will
setup a Sourcegraph instance (single replica) on a Google Container Engine
cluster. With small modification to the configuration below you can setup
Sourcegraph to run on any Kubernetes infrastructure.

# Volumes

A Sourcegraph Pod must mount two volumes: one to persist server configuration
and another to persist repositories and other Sourcegraph metadata.

On Google infrastructure, create a Persistent Disk for each volume:

```
gcloud compute disks create --size=300GB sourcegraph-data
gcloud compute disks create --size=10GB sourcegraph-config
```

Choose a disk size for `sourcegraph-data` based on your repository storage
requirements.

# Service

Create a load balancer with a public IP address.

First, create a `service.yml` file with the following contents:

```
apiVersion: v1
kind: Service
metadata:
  name: sourcegraph
spec:
  selector:
    app: sourcegraph
  ports:
  - name: http
    protocol: TCP
    port: 80
    targetPort: 80
  - name: https
    protocol: TCP
    port: 443
    targetPort: 443
  type: LoadBalancer
```

Then, create your Service.

```
kubectl create -f service.yml
```

# Replication Controller

Create a Replication Controller to pull Sourcegraph's docker image
and mount your volumes.

First, create a `rc.yml` file with the following contents:

```
apiVersion: v1
kind: ReplicationController
metadata:
  name: sourcegraph
spec:
  replicas: 1
  selector:
    app: sourcegraph
  template:
    metadata:
      labels:
        app: sourcegraph
    spec:
      containers:
      - name: src
        image: sourcegraph/sourcegraph:latest
        volumeMounts:
        - name: data
          mountPath: /etc/sourcegraph
        - name: config
          mountPath: /home/sourcegraph/.sourcegraph
        ports:
        - containerPort: 80
          protocol: TCP
        ports:
        - containerPort: 443
          protocol: TCP
      volumes:
      - name: data
        gcePersistentDisk:
          pdName: sourcegraph-data
          fsType: "ext4"
      - name: config
        gcePersistentDisk:
          pdName: sourcegraph-config
          fsType: "ext4"
      restartPolicy: Always
```

Then, create your Replication Controller:

```
kubectl create -f rc.yml
```

NOTE: if your disks aren't on Google infrastructure, you'll need to
modify the `volumes` property of the `rc.yml` to use something other than
`gcePersistentDisk`.

# Sanity Check

You should now have a running, yet unconfigured, instance of Sourcegraph!

To verify this, get the IP address of the load balancer. This IP
will be labeled as *LoadBalancer Ingres* when running the following command:

```
kubectl describe services/sourcegraph
```

Visit the IP address in your browser. You are ready to use Sourcegraph!

# Configure Sourcegraph

To configure your Sourcegraph server, first get the name of the Pod
that Sourcegraph is running on:

```
kubectl get pods
```

Choose the Pod prefixed by the name of the Replication Controller,
and edit your Sourcegraph configuration:

```
kubectl exec <Pod Name> -i -t -- vi /etc/sourcegraph/config.ini
```

After editing the configuration, restart the Sourcegraph instance
by deleting the Pod:

```
kubectl delete pods/<Pod Name>
```

{{< ads_conversion >}}
