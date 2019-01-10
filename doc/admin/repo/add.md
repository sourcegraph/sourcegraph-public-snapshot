# Add repositories (from code hosts) to Sourcegraph

- [Add repositories from GitHub or GitHub Enterprise](../../integration/github.md)
- [Add repositories from GitLab](../../integration/gitlab.md)
- [Add repositories from Bitbucket Server](../../integration/bitbucket_server.md)
- [Add repositories from AWS CodeCommit](../../integration/aws_codecommit.md)
- [Add repositories from Phabricator](../../integration/phabricator.md)
- [Add repositories from the local disk](add_from_local_disk.md)

## Troubleshooting

If your repositories are not showing up:

- On single-node deployments, check the logs from the `sourcegraph/server` Docker container for error messages related to communication with your code host's API.
- On Kubernetes cluster deployments, check the logs from the `repo-updater` pod.
- Check the site admin **Repositories** page on Sourcegraph (and ensure you're logged in as an admin).

If your repositories are showing up but are not cloning or updating from the original Git repository:

- Go to the repository's **Mirroring** settings page and inspect the **Check connection** logs.
