# Administration FAQ

## How do I expose my Sourcegraph instance to a different host port when running locally?

Change the `docker` `--publish` argument to make it listen on the specific interface and port on your host machine. For example, `docker run ... --publish 0.0.0.0:80:7080 ...` would make it accessible on port 80 of your machine. For more information, see "[Publish or expose port](https://docs.docker.com/engine/reference/commandline/run/#publish-or-expose-port--p---expose)" in the Docker documentation.

The other option is to deploy and run Sourcegraph on a cloud provider. For an example, see documentation to [deploy to Google Cloud](install/docker/google_cloud.md).

## How do I access the Sourcegraph database?

> NOTE: To execute an SQL query against the database without first creating an interactive session (as below), append `--command "SELECT * FROM users;"` to the `docker container exec` command.

### For single-node deployments (`sourcegraph/server`)

Get the Docker container ID for Sourcegraph:

```bash
docker ps
CONTAINER ID        IMAGE
d039ec989761        sourcegraph/server:VERSION
```

Open a PostgreSQL interactive terminal:

```bash
docker container exec -it d039ec989761 psql -U postgres sourcegraph
```

Run your SQL query:

```sql
SELECT * FROM users;
```

### For Kubernetes cluster deployments

Get the id of one `pgsql` Pod:

```bash
kubectl get pods -l app=pgsql
NAME                     READY     STATUS    RESTARTS   AGE
pgsql-76a4bfcd64-rt4cn   2/2       Running   0          19m
```

Open a PostgreSQL interactive terminal:

```bash
kubectl exec -it pgsql-76a4bfcd64-rt4cn -- psql -U sg
```

Run your SQL query:

```sql
SELECT * FROM users;
```

## How does Sourcegraph store repositories on disk?

Sourcegraph stores bare Git repositories (without a working tree), which is a complete mirror of the repository on your code host.

If you are keen for more details on what bare Git repositories are, [check out this discussion on StackOverflow](https://stackoverflow.com/q/5540883).

The directories should contain just a few files and directories, namely: HEAD, config, description, hooks, info, objects, packed-refs, refs

## Does Sourcegraph support svn?

Sourcegraph natively supports Git repositories, but Subversion repositories can be indexed through
`git svn` or other svn-to-git translation tools. Here is a rough outline of how to make a Subversion
repository available on Sourcegraph:

1. Convert the Subversion repository to Git using [`git svn`](https://git-scm.com/docs/git-svn). For
   larger repositories, `git svn` may take
   awhile. [`svn2git`](https://github.com/svn-all-fast-export/svn2git) is a faster alternative that
   can be run as a Docker container.
1. Once the Subversion repository is converted to a Git repository, push it to a Git code host that
   is accessible by Sourcegraph. If you do not have a Git code host, you can [set up a simple Git
   server](https://www.linux.com/tutorials/how-run-your-own-git-server/) or run
   [GitLab](https://about.gitlab.com/install/), which is free and open-source.
1. Connect Sourcegraph to your Git code host and [configure
   it](https://docs.sourcegraph.com/admin/external_service) to index your converted Subversion
   repositories.

## Troubleshooting

Content moved to a [dedicated troubleshooting page](troubleshooting.md).
