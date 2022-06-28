# Configure Sourcegraph on Elastic Kubernetes Service (EKS)

To install Sourcegraph on AWS Elastic Kubernetes Service, you must deploy onto a supported machine type and use a persistent standard disk or a persistent SSD.

## Prerequisites {#eks-prerequisites}

- You need to have an EKS cluster (>=1.19) with the following addons enabled:
  - [AWS Load Balancer Controller](https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html)
  - [AWS EBS CSI driver](https://docs.aws.amazon.com/eks/latest/userguide/managing-ebs-csi.html)
  
> [Learn more](https://kubernetes.github.io/ingress-nginx/) about deploying your own Ingress Controller instead of the ALB Ingress Controller.

- You need to have an account with sufficient access equivalent to the `cluster-admin` ClusterRole.
- You need to be able to connect to your cluster (via either the console or the command line using `eksctl`) and ensure the cluster is up and running. You should see several `ready` nodes listed when you run: `kubectl get nodes`.
- You need to have the [Helm CLI](https://helm.sh/docs/intro/install/) installed and run the following command to link to the Sourcegraph helm repository. This should be ran on the machine used to interact with your cluster:

```sh
helm repo add sourcegraph https://helm.sourcegraph.com/release
```

## Hardware and Service Requirements

Use the [resource estimator](../resource_estimator.md) to determine the resource requirements for your environment. You will use this information to set up the instance and configure the override file in the steps below.

## Steps {#eks-steps}

### Create Override File & Add Configurations

Create an override file and add any configuration override settings you need. See the [configuration](#configuration) documentation for more information on override files and the options for configurations.

We recommend adding the following values into your override file to configure Ingress to use [AWS Load Balancer Controller](https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html), to expose Sourcegraph publicly on a domain of your choosing, and to configure the Storage Class to use [AWS EBS CSI driver](https://docs.aws.amazon.com/eks/latest/userguide/managing-ebs-csi.html). For an example, see [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/aws/override.yaml).

Uncomment the `provisioner` that your Amazon EKS cluster implements.

<!--[override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/aws/override.yaml)-->
```yaml
frontend:
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: alb # aws load balancer controller ingressClass name
      # additional aws alb ingress controller supported annotations
      # ...
    # replace with your actual domain
    host: sourcegraph.company.com

storageClass:
  create: true
  type: gp2 # This configures SSDs (recommended).
#  provisioner: ebs.csi.aws.com # use this provisioner if using the self-managed Amazon EBS Container Storage Interface driver 
#  provisioner: kubernetes.io/aws-ebs # use this provisioner if using the Amazon EKS add-on
  volumeBindingMode: WaitForFirstConsumer
  reclaimPolicy: Retain
```

> ℹ️ You can review the changes using one of the [three mechanisms](./helm#reviewing-changes) to assess the customizations made. This is not required, but may be useful the first time you deploy Sourcegraph.

### Install the Sourcegraph Helm chart

Install the Sourcegraph Helm chart by running the following command:

```sh
helm upgrade --install --values ./override.yaml --version 3.41.0 sourcegraph sourcegraph/sourcegraph
```

It will take some time for the load balancer to be fully ready. Use the following command to check on the status and obtain the load balancer address once available:

```sh
kubectl describe ingress sourcegraph-frontend
```

### Create a DNS Record

Once you have obtained the allocated address of the load balancer, you should create a DNS record for the `sourcegraph.company.com` domain that resolves to the load balancer address.

It is recommended to enable TLS and configure a certificate on your load balancer. We recommend using an [AWS-managed certificate](https://docs.aws.amazon.com/acm/latest/userguide/acm-overview.html) and add the following annotations to Ingress:

```yaml
frontend:
  ingress:
    annotations:
      kubernetes.io/ingress.class: alb
      # ARN of the AWS-managed TLS certificate
      alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-west-2:xxxxx:certificate/xxxxxxx
```

### Validate the deployment

Sourcegraph should now be available via the address set. 

Navigate to the URL in your browser to ensure you now have access to the Sourcegraph UI to create the initial administrator account.

### Sourcegraph Configuration

At this stage the deployment is considered to be complete. You are now ready to configure your Sourcegraph instance (site configuration, code host configuration, search configuration etc). Please see our [Configuring Sourcegraph](../../config/index.md) documentation for guidance.

## References {#eks-references}

- [Enable TLS with AWS-managed certificate](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations/#ssl)
- [Supported AWS load balancer annotations](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations)