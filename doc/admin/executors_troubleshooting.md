# Troubleshooting Executors
This page compiles a list of common troubleshooting steps found during development and administration of executors.


## Disabling the auto-deletion of Executor VMs
> NOTE: These instructions are for users using the VMs deployed via the [Terraform Modules](https://docs.sourcegraph.com/admin/deploy_executors_terraform)

The Executor host VMs are configured to automatically tear themselves down once all jobs in the queue are completed. This is very inconvenient when trying to debug issues in the executor configuration or connections. To prevent the VMs from automatically stopping:
1. `ssh` into the VM
1. `sudo su` to become the `root` user
1. Remove (or rename) the `/shutdown_executor.sh` file

The VM should now persist after all jobs are satisfied.

## Creating a Debug Firecracker VM
To create a temporary Firecracker VM for debugging purposes:

> NOTE: if the host VM is provisioned with the [Sourcegraph terraform modules](https://docs.sourcegraph.com/admin/deploy_executors_terraform), the VMs may be configured to sotp automatically. Refer to [Disabling the auto-deletion of Executor VMs](#disabling-the-auto-deletion-of-executor-vms) for information to prevent this.

1. `ssh` into the host VM
1. `sudo su` to become the `root` user
1. `systemctl stop executor` to stop the `executor` service
1. `export $(cat /etc/systemd/system/executor.env | xargs)` to load the executor environment into your shell
1. Run `executor test-vm` to generate a test firecracker VM. The command will output a line like:
  ```
  Success! Connect to the VM using
    $ ignite attach executor-test-vm-0160f53f-e765-4481-a81e-aa3c704d07bd
  ```
1. Execute the generated `ignite attach <vm>` command to gain a shell to the Firecracker VM

## Recreating a Firecracker VM 
If a server-side batch change fails unexpectedly, it's possible to recreate the generated Firecracker VM from the batch change execution.

> NOTE: if the host VM is provisioned with the [Sourcegraph terraform modules](https://docs.sourcegraph.com/admin/deploy_executors_terraform), the VMs may be configured to stop automatically. Refer to [Disabling the auto-deletion of Executor VMs](#disabling-the-auto-deletion-of-executor-vms) for information to prevent this.

1. Navigate to the failed execution page of the Batch Change
1. Select a failed Workspace on the left and click the `Diagnostics` link on the right pane
1. In the modal, expand the `Setup` step by clicking the text or the expansion arrow on the right
1. Copy the command from the  final step of `Setup` starting with `ignite run` 

1. `ssh` into the host VM
1. `sudo su` to become the `root` user
1. `systemctl stop executor` to stop the `executor` service
1. `export $(cat /etc/systemd/system/executor.env | xargs)` to load the executor environment into your shell
1. Paste in the command copied from the batch change. You may need to remove the `--copy-files` and `--volumes` directives as those volumes and files may not exist on the VM any longer. Surround the `--kernal-args` arguments in quotes as well.
1. Execute the command and wait for the VM to start
1. Run `ignite ps` to list all currently running VMs
1. Run `ignite attach <vm id>` to get a shell to the running VM

