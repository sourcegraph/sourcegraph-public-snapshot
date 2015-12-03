+++
title = "GitHub.com"
+++

## Public repositories

To mirror public GitHub.com repositories on a Sourcegraph instance, run this command:

```
src repo create -m --clone-url https://github.com/mycompany/project <repo-name>
```

## Private repositories

Sourcegraph can import private GitHub.com repositories, enabling a limited set of Sourcegraph features for the externally hosted repository.
Navigate to `http://src.mycompany.com/~USERNAME/.settings/integrations` and follow the instructions to create and add a GitHub personal access
token to your Sourcegraph instance. This will fetch the list of private repositories. Enable a repository to mirror the upstream GitHub repository
on Sourcegraph.
