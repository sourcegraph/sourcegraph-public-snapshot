+++
title = "Git mirroring"
description = "Make Sourcegraph read from or write to an external repository"
+++

Sourcegraph works great as the system of record, but there are some situations in which you will want to mirror your repositories hosted on Sourcegraph to a Git repository hosted somewhere else. With configuration, Sourcegraph can automatically push changes to an external code host each time someone pushes changes to Sourcegraph.

Likewise, there are situations in which you will want to to mirror your externally hosted Git repositories on Sourcegraph. With this configuration, Sourcegraph will poll the external repository for changes on a short interval.

# Mirroring *from* externally hosted repositories

## Public repository

Let's say you have a repository hosted on an `host.com/my/repo` and with public read access and want to
have this repository automatically mirrored on your Sourcegraph server. Use the `src` CLI to create
the mirror:

```
src repo create -m --clone-url=http://host.com/my/repo <repo-name>
```

## Private repository

## 1. Authenticating with SSH keys

GitHub, Bitbucket, and other git hosting services often allow users to configure their hosts to authenticate using SSH keys (See [here](https://help.github.com/articles/generating-ssh-keys/) for GitHub's instructions, and [here](https://confluence.atlassian.com/bitbucket/set-up-ssh-for-git-728138079.html) for Bitbucket's instructions on how to set this up).

Determine the SSH link for the private repository you would like to clone. It typically has the form of `git@bitbucket.org:userorganization/test_repository.git`, and can be accessed on your GitHub or Bitbucket repository page (you may have to select SSH from the standard clone widget, HTTPS is selected by default).


First, check that the git clone command functions with this SSH git clone URL:
```
git clone git@bitbucket.org:userorganization/test_repository.git
```

Ensure that the clone command completely successfully. You can now configure Sourcegraph to mirror your private repository:
```
src repo create -m --clone-url=git@bitbucket.org:userorganization/test_repository.git <repo-name>
```

## 2. Authenticating over HTTPS

Alternatively, you can mirror a private repository by using an HTTPS git clone link. GitHub and Bitbucket make these links available on repository pages, they typically look like this:
`https://user@bitbucket.org/userorganization/test_repository.git`.

First, append your password to the link, such as: `https://user:password@bitbucket.org/userorganization/test_repository.git`.

Next, check that the git clone command functions with this URL:
```
git clone https://user:password@bitbucket.org/userorganization/test_repository.git
```

Ensure that the clone command completely successfully. You can now configure Sourcegraph to mirror your private repository:
```
src repo create -m --clone-url=https://user:password@bitbucket.org/userorganization/test_repository.git <repo-name>
```

## 3. Using GitHub access tokens
Sourcegraph currently supports importing private repositories from GitHub.com and GitHub Enterprise.
Follow instructions for [GitHub.com]({{< relref "integrations/GitHub.md" >}}) or
[GitHub Enterprise]({{< relref "integrations/GitHub_Enterprise.md" >}}).

# Mirroring *to* externally hosted repositories

Let's say you have a repository hosted on Sourcegraph (`src.example.com/my/repo`) and want to have this repository automatically mirrored on `github.com/my/repo`. Here's what you need to do:

## 1. Ensure you can push changes

First, get shell access to your Sourcegraph server (`src.example.com`) and ensure that you can push changes to `github.com/my/repo` _without specifying a username/password on the command line_ (you can do this by configuring your SSH keys as you would normally).

  - If you skip this step, Sourcegraph won't be able to push to the remote repository.

## 2. Add config option

Now you'll need to either use the `src serve --fs.git-repo-mirror` CLI flag or add this to your configuration file:

```
[serve.Local filesystem storage (fs store)]
GitRepoMirror = <LocalRepoURI>:<GitRemoteURL>
```

- More than one repo can be mirrored by simply adding a comma.
- Note: If you installed Sourcegraph using one of the standard distribution or cloud provider packages,
Sourcegraph will run with configuration found at `/etc/sourcegraph/config.ini`.

Example:

```
[serve.Local filesystem storage (fs store)]
GitRepoMirror = my/repo:git@github.com:my/repo,other/hacks:git@github.com:me/hacks
```

Would cause any pushes to `src.example.com/my/repo` to be mirrored to `git@github.com/my/repo` and likewise any changes to `src.example.com/other/hacks` would be mirrored to `git@github.com:me/hacks`.

## 3. Restart Sourcegraph

After updating your config, restart the `src` process:

```
sudo restart src
```

## 4. Test it out!

Now that everything is configured, simply push any change to your Sourcegraph repository and watch as it is mirrored to the remote Git URL you specified!