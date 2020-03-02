# Add repositories already cloned to disk

You can also add repositories to Sourcegraph that are already cloned to disk on the host machine. This is useful for repositories requiring non-standard authentication to clone, or very large repositories on which cloning exceeds the resources available to the Docker container. (It is not intended for individual users who want to set up a personal Sourcegraph instance just for searching code on their own local disk; we recommend just using a CLI tool such as ripgrep instead.) 

> WARNING: Sourcegraph will alter the contents and structure of files under `/var/opt/sourcegraph` (Sourcegraph’s data volume inside the container), so do not mount repositories in use by other processes under that directory.

The steps documented here are intended for Sourcegraph instances running on a single node. The general process also applies for clustered deployments of Sourcegraph to Kubernetes, but you need to perform these steps on the underlying node hosting the `gitserver` pod, or on the persistent volume used by the `gitserver` deployment.

If you're using the default `--volume $HOME/.sourcegraph/data:/var/opt/sourcegraph` argument to run the `sourcegraph/server` Docker image, and the repository you want to add is named `github.com/my/repo`, then follow these steps:

1.  If Sourcegraph is running, ensure the repository is disabled so it doesn't attempt to clone it.

1.  On the host machine, ensure that a bare Git clone of the repository exists at `$HOME/.sourcegraph/data/repos/github.com/my/repo/.git`.

    To create a new clone given its clone URL:

    ```
    git clone --mirror YOUR-REPOSITORY-CLONE-URL $HOME/.sourcegraph/data/repos/github.com/my/repo/.git
    ```

    Or, as an optimization, you can reuse an existing local clone to avoid needing to fetch all the repository data again:

    ```
    git clone --mirror --reference PATH-TO-YOUR-EXISTING-LOCAL-CLONE --dissociate YOUR-REPOSITORY-CLONE-URL $HOME/.sourcegraph/data/repos/github.com/my/repo/.git
    ```

1.  Ensure that the code host of the added repository is configured as an [external service](../external_service/index.md).
1.  Enable the repository on the site admin repositories page.

If this repository exists on a code host that Sourcegraph directly integrates with, then use that code host's configuration (as described in the [external service documentation](../external_service/index.md)). After updating the external service configuration, if you used the correct repository path, Sourcegraph will detect and reuse the existing clone. (For example, if you're working with a repository on GitHub.com, ensure that the repository path name you used is of the form `github.com/my/repo`.)
