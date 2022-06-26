# Configure Sourcegraph on Elastic Kubernetes Service (EKS)

## Prerequisites {#eks-prerequisites}

1. You need to have a EKS cluster (>=1.19) with the following addons enabled:
   - [AWS Load Balancer Controller](https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html)
   - [AWS EBS CSI driver](https://docs.aws.amazon.com/eks/latest/userguide/managing-ebs-csi.html)
> You may consider deploying your own Ingress Controller instead of the ALB Ingress Controller, [learn more](https://kubernetes.github.io/ingress-nginx/)
1. Your account should have sufficient access equivalent to the `cluster-admin` ClusterRole.
1. Connect to your cluster (via either the console or the command line using `eksctl`) and ensure the cluster is up and running using: `kubectl get nodes` (several `ready` nodes should be listed)
1. Have the [Helm CLI](https://helm.sh/docs/intro/install/) installed and run the following command to link to the Sourcegraph helm repository (on the machine used to interact with your cluster):

```sh
helm repo add sourcegraph https://helm.sourcegraph.com/release
```

## Steps {#eks-steps}

**1** – Create your override file and add in any configuration override settings you need - see [configuration](#configuration) for more information on override files and the options around what can be configured.

We recommend adding the following values into your override file to configure Ingress to use [AWS Load Balancer Controller] to expose Sourcegraph publicly on a domain of your choosing, and to configure the Storage Class to use [AWS EBS CSI driver]. For an example, see [override.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/aws/override.yaml).

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
> ℹ️ Optionally, you can review the changes using one of [three mechanisms](#reviewing-changes) that can be used to assess the customizations made. This is not required, but may be useful the first time you deploy Sourcegraph, for peace of mind.

**2** – Install the chart

```sh
helm upgrade --install --values ./override.yaml --version 3.41.0 sourcegraph sourcegraph/sourcegraph
```

It will take some time for the load balancer to be fully ready, use the following to check on the status and obtain the load balancer address (once available):

```sh
kubectl describe ingress sourcegraph-frontend
```

**3** – Upon obtaining the allocated address of the load balancer, you should create a DNS record for the `sourcegraph.company.com` domain that resolves to the load balancer address.

It is recommended to enable TLS and configure a certificate properly on your load balancer. You may consider using an [AWS-managed certificate](https://docs.aws.amazon.com/acm/latest/userguide/acm-overview.html) and add the following annotations to Ingress.

```yaml
frontend:
  ingress:
    annotations:
      kubernetes.io/ingress.class: alb
      # ARN of the AWS-managed TLS certificate
      alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-west-2:xxxxx:certificate/xxxxxxx
```

**4** – Validate the deployment
Sourcegraph should now be available via the address set.
Browsing to the url should now provide access to the Sourcegraph UI to create the initial administrator account.

**5** – Further configuration

Now the deployment is complete, more information on configuring the Sourcegraph application can be found here:
[Configuring Sourcegraph](../../config/index.md)

## References {#eks-references}

- [Enable TLS with AWS-managed certificate](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations/#ssl)
- [Supported AWS load balancer annotations](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations)