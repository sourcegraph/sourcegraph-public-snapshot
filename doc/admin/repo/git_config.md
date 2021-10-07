# Custom git config

Sourcegraph supports customising [git-config](https://git-scm.com/docs/git-config) and [ssh_config](https://linux.die.net/man/5/ssh_config) for adjusting the behaviour of git. Sourcegraph will read these from the standard locations.

This guide documents how to configure git-config. To set up SSH and authentication for repositories, see [Repository authentication](auth.md).

- [Sourcegraph with Docker Compose](../install/docker-compose/index.md): refer to the Docker Compose guide to [set up custom Git configuration and authentication](../install/docker-compose/operations.md#git-configuration-and-authentication), which can be adapted to additionally set `/etc/gitconfig`.
- [Sourcegraph with Kubernetes](../install/docker-compose/index.md): refer to the Kubernetes guide to [configure repository cloning via SSH](../install/kubernetes/configure.md#configure-repository-cloning-via-ssh), which can be adapted to additionally set `/etc/gitconfig`.
- [Single-container Sourcegraph](../install/docker-compose/index.md): refer to the single-container guide to [set up custom Git configuration and authentication](../install/docker-compose/operations.md#git-configuration-and-authentication).
- [Pure-docker Sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/pure-docker): refer to the Pure-Docker guide to [configure SSH cloning](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/pure-docker/README.md#configuring-ssh-cloning), which can be adapted to additionally set `/etc/gitconfig`.

## Example: alternate clone URL for repos

Some sites put an HTTPS or SSH proxy in front of their code host to reduce load. Some sites also use a service like AWS CodeCommit to do the same. In these cases, the repos still should be treated as being repos on the original code host, not the proxy site.

For example, I have a GitHub repo `github.com/foo/bar`. I want Sourcegraph to clone it using the URL https://cloneproxy.example.com/foo/bar.git. But I still want the "Go to GitHub repository" button, etc., to take me to https://github.com/foo/bar. You can use the git configuration [`insteadOf`](https://git-scm.com/docs/git-config#Documentation/git-config.txt-urlltbasegtinsteadOf):

``` ini
# ~/.gitconfig or /etc/gitconfig
[url "https://cloneproxy.example.com"]
  insteadOf = https://github.com
```

If you are [cloning via SSH](./auth.md), you can also achieve this with an SSH configuration:

```
# ~/.ssh/config
Host github.com
  Hostname cloneproxy.example.com
```
