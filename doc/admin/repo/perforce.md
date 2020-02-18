# Using Perforce repositories with Sourcegraph

You can use [Perforce](https://perforce.com) repositories with Sourcegraph by using the [git p4](https://git-scm.com/docs/git-p4) adapter, which creates an equivalent Git repository from a Perforce repository. Sourcegraph doesn't yet support Perforce repositories natively.

Screenshot of using Sourcegraph for code navigation in a Perforce repository:

![Viewing a Perforce repository on Sourcegraph](https://storage.googleapis.com/sourcegraph-assets/git-p4-example.png)

## src-expose

We have an experimental alternative for importing Perforce code into Sourcegraph: [src-expose](../external_service/other.md#experimental-src-expose).

## Instructions

### Prerequisites

- Git
- Perforce `p4` CLI configured to access your Perforce repository
- `git p4` (see "[Adding `git p4` to an existing install](https://git.wiki.kernel.org/index.php/GitP4#Adding_git-p4_to_an_existing_install)")
- A Git host (such as GitHub or GitLab) where you can push new Git repositories

### Create an equivalent Git repository from a Perforce repository

For each Perforce repository you want to use with Sourcegraph, follow these steps:

1. Create a local Git repository with the contents of your Perforce repository: `git p4 clone //DEPOT/PATH@all` (replace `//DEPOT/PATH` with the Perforce repository path).
1. `cd PATH` to enter the directory of the new local Git repository.
1. Create a new Git repository on your Git host (such as GitHub or GitLab) for this repository.
1. Add the repository you just created on the Git host as a remote: `git remote add origin https://git-host.example.com/my/repo.git`
1. Push the repository to the Git host: `git push -u origin master`

#### Updating Perforce repositories

To update the repository after new Perforce commits are made, run `git p4 sync && git push` in the local repository directory. Sourcegraph does not yet automatically sync repositories from Perforce, so you must do this manually or script it yourself.

### Add the converted Perforce repositories from your Git host to Sourcegraph

The repositories you created on your Git host are normal Git repositories, so you can [add the repositories to Sourcegraph](index.md) as you would any other Git repositories.

## Known issues

We intend to improve Sourcegraph's Perforce support in the future. Please [file an issue](https://github.com/sourcegraph/sourcegraph/issues) to help us prioritize any specific improvements you'd like to see.

- Sourcegraph was initially built for Git repositories only, so it exposes Git concepts that are meaningless for converted Perforce repositories, such as the commit SHA, branches, and tags.
- There is no automatic updating of converted repositories when new Perforce commits are made. See the "[Updating Perforce repositories](#updating-perforce-repositories)" section for manual steps.
- The commit messages for a Perforce repository converted to a Git repository have an extra line at the end with Perforce information, such as `[git-p4: depot-paths = "//guest/acme_org/myproject/": change = 12345]`.
