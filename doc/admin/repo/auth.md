# Repositories that need HTTP(S) or SSH authentication

If authentication is required to `git clone` a repository then you must provide credentials to the container.

First, ensure your **Site admin > Manage repositories** code host configuration is configured to use SSH. For example, by setting the `gitURLType` field to "ssh". Alternatively, you may use the "Generic Git host" code host type, which allows you to directly specify Git repository URLs for cloning.

Then, follow the directions below depending on your deployment type.

## For single-node deployments (`sourcegraph/server`)

### SSH authentication (config, keys, `known_hosts`)

The container consults its own file system (in the standard locations) for SSH configuration, private keys, and `known_hosts`. Upon container start, it copies all files from `/etc/sourcegraph/ssh` into its own `$HOME/.ssh` directory.

To provide SSH authentication configuration to the container, assuming you're using the default `--volume $HOME/.sourcegraph/config:/etc/sourcegraph`, follow these steps:

1. Create files at `$HOME/.sourcegraph/config/ssh/config`, `$HOME/.sourcegraph/config/ssh/known_hosts`, etc., on the host machine as desired to configure SSH.
1. Start (or restart) the container.

To configure the container to use the same SSH as your user account on the host machine, you can also run `cp -R $HOME/.ssh $HOME/.sourcegraph/config/ssh`.

### HTTP(S) authentication via netrc

The easiest way to specify HTTP(S) authentication for repositories is to include the username and password in the clone URL itself, such as `https://user:password@example.com/my/repo`. These credentials won't be displayed to non-admin users.

Otherwise, the container consults the `$HOME/.netrc` files on its own file system for HTTP(S) authentication. The `.netrc` file is a standard way to specify authentication used to connect to external hosts.

To provide HTTP(S) authentication, assuming you're using the default `--volume $HOME/.sourcegraph/config:/etc/sourcegraph`, follow these steps:

1. Create a file at `$HOME/.sourcegraph/config/netrc` on the host machine that contains lines of the form `machine example.com login alice password mypassword` (replacing `example.com`, `alice`, and `mypassword` with the actual values).
1. Start (or restart) the container.

## For Docker Compose deployments

## SSH authentication (config, keys, `known_hosts`)

Provide your `gitserver` instance with your SSH / Git configuration (e.g. `.ssh/config`, `.ssh/id_rsa`, `.ssh/id_rsa.pub`, and `.ssh/known_hosts`--but you can also provide other files like `.netrc`, `.gitconfig`, etc. if needed) by mounting a directory that contains this configuration into the `gitserver` container.

For example, in the `gitserver-0` container configuration in your docker-compose.yaml file, add the second volume listed below, replacing `~/path/on/host/` with the path on the host machine to the `.ssh` directory:

```
gitserver-0:
  container_name: gitserver-0
  ...
  volumes:
    - 'gitserver-0:/data/repos'
    - '~/path/on/host/.ssh:/home/sourcegraph/.ssh'
  ...
```

>NOTE: The permission of your SSH / Git configuration must be set to be readable by the user in the `gitserver` container. For example, run `chmod -v -R 600 ~/path/to/.ssh` in the folder on the host machine.

See [Custom git or ssh config docs](https://docs.sourcegraph.com/admin/repo/custom_git_or_ssh_config#setting-configuration) on setting custom configuration 

### Troubleshooting

#### What should be included in my config file?

We recommend adding the `StrictHostKeyChecking no` and `AddKeysToAgent yes` flags to prevent the need to give permission interactively when cloning from a new host.

```yaml
Host *
  StrictHostKeyChecking no
  AddKeysToAgent yes
```


#### Error: `Host key verification failed`
This indicates an invalid key is being used. You can confirm the error by cloning inside the gitserver directly. For example:
```bash
docker exec -it gitserver-0 sh
cd data/repos/<CODE-HOST>/<REPO-OWNER>
git clone <SSH-URL>
```

#### Error: `Bad owner or permissions on /home/sourcegraph/.ssh/<YOUR-CONFIG-FILE>`
This indicates the container is having trouble reading the configuration files due to permission / owner issues.
The permission and ownership settings inside your `.ssh/` directory should look similar to:
```bash
$ ls -al #command to display list of file with detailed information
total 20
drwxr-xr-x    6 sourcegr sourcegr       192 May 12 19:54 .
drwxr-sr-x    1 sourcegr sourcegr      4096 May 12 19:43 ..
-rw-------    1 sourcegr sourcegr        34 May 12 19:22 config
-rw-------    1 sourcegr sourcegr       411 May 12 18:52 id_ed25519
-rw-------    1 sourcegr sourcegr        98 May 12 18:52 id_ed25519.pub
-rw-------    1 sourcegr sourcegr       799 May 12 19:54 known_hosts
```
Solution:
- Inside the `.ssh` directory on the Host Machine:
  - Permission on all files must be set to `600`, and `700` for the directory itself.
  - Files must be owned by a user who has access to the docker container. This can be done via `sudo chown -v -R $USER:$GROUP` (the user may need to set these values).
- (OPTIONAL: Please read note below) Inside the `/home/sourcegraph/` directory on Docker Compose:
  - Permission on all files must be set to `600`, and `700` for the directory itself.
  - Files must be owned by the root user, which is `sourcegraph` by default. This can be done via `sudo chown -v -R $USER:$GROUP` (the user may need to set these values).

>NOTE: Once the volume for the configuration files has been mounted, both the `/ssh` directory on the host machine and docker will be synced and changes within one directory will be reflected by the other one. Consquently, you will only need to make the changes within one directory.

#### Error: `Permissions 0644 for '/home/sourcegraph/.ssh/<YOUR-PRIVATE-KEY-FILE>' are too open`
This indicates the permission on your private key file is accessible by users other than the file owner. Setting the file permission to 600 resolves the issue.

### HTTP(S) authentication via netrc

The easiest way to specify HTTP(S) authentication for repositories is to include the username and password in the clone URL itself, such as `https://user:password@example.com/my/repo`. These credentials won't be displayed to non-admin users.

Otherwise, follow the steps above for mounting SSH configuration to mount a host directory containing the desired `.netrc` file to `/home/sourcegraph/` in the `gitserver` container.

## For Kubernetes cluster deployments

See "[Configure repository cloning via SSH
](../install/kubernetes/configure.md#configure-repository-cloning-via-ssh)" in the Kubernetes cluster administrator guide.

## For pure-Docker cluster deployments

See "[Configuring SSH cloning](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/README.md#configuring-ssh-cloning)" in the Pure-Docker Sourcegraph cluster deployment reference.
