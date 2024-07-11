# Appliance

Appliance provides a platform for configuration and automation of Sourcegraph deployments and administration in a Kubernetes environment. This allows users to easily setup and configure Sourcegraph in their environment as well as more easily manage administration tasks such as upgrades.

---

## Architecture

Appliance runs as a standard Kubernetes Deployment and utilizes Kubernetes [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) in order to manage deployment and administration tasks.

## Own

For more information or for help, see the [Release Team](https://handbook.sourcegraph.com/departments/engineering/teams/release/).

## Development

You can kick the tires on the appliance version by running:

```
go run ./cmd/appliance
```

[`config.go`](./shared/config.go) is the source of truth on appliance
configuration. Most of the variables there are optional, except for:

- `APPLIANCE_VERSION`: while this does have a default that does not need to be
  overridden in production, development builds that lack the link-time injected
  version information will need to set this. Set it to the latest version of
  Sourcegraph that you want to be offered.

You might want to override the listen addresses to localhost-only, in order to
avoid macos firewall popups.

The appliance doesn't care if it's running inside or out of the k8s cluster it's
provisioning resources into. It does a well-known k8s config dance to try to
load in-cluster config (from a k8s ServiceAccount token), falling back on
looking for a kubeconfig on the host

If you have some kubernetes running (e.g. minikube, docker desktop), and your
default context is set in ~/.kube/config, the appliance will build a k8s client
using that kubeconfig, and everything should "just work".

You must set an admin password, e.g:

```
SG_APPLIANCE_PW=$(pwgen -s 40 1)
echo -e "Your Sourcegraph appliance password is:\n\n${SG_APPLIANCE_PW}\n"
kubectl -n test create secret generic appliance-password --from-literal password="${SG_APPLIANCE_PW}"
```

On first boot the appliance will hash that password, transpose it to another
backing secret, and delete the secret you just created.

To reset the appliance password:

```
kubectl -n test delete secret appliance-data
```

And then create the password again as per the above instructions.

See [`development.md`](../..internal/appliance/development.md) for more
information, including about automated testing.
