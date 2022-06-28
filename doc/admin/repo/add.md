# Add repositories (from code hosts) to Sourcegraph

- [Add repositories from a code host](../external_service/index.md) (GitHub, GitLab, Bitbucket Server, Bitbucket Data Center, AWS CodeCommit, Phabricator, or Gitolite)
- [Add repositories by Git clone URLs](../external_service/other.md)
- [Add repositories from non-Git code hosts](../external_service/non-git.md)
  - [Add Perforce repositories](perforce.md)
- [Pre-load repositories from the local disk](pre_load_from_local_disk.md)

## Troubleshooting

If your repositories are not showing up, check the site admin **Repositories** page on Sourcegraph (and ensure you're logged in as an admin).
If nothing informative is visible there, check for error messages related to communication with your code host's API in the logs from:

- [Docker Compose](../deploy/docker-compose/index.md) and [Kubernetes](../deploy/kubernetes/index.md): the logs from the `repo-updater` container/pod
- [Single-container](../deploy/docker-single-container/index.md): the `sourcegraph/server` Docker container

If your repositories are showing up but are not cloning or updating from the original Git repository:

- Go to the repository's **Mirroring** settings page and inspect the **Check connection** logs.
