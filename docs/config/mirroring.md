+++
title = "Git mirroring"
description = "Make Sourcegraph read from or write to an external repository"
+++

Sourcegraph is designed to function as a standalone git server.
You may also configure Sourcegraph to read from or write to an externally
hosted repository (e.g. on GitHub or Bitbucket).

# Mirroring *from* externally hosted repositories

## Public repository

If your repository may be read without authentication, use the `src` CLI to create a mirror:

```
src repo create -m --clone-url=http://host.com/my/repo <repo-name>
```

## Private repository

### 1. Authenticating with SSH keys

GitHub, Bitbucket, and other git hosting services often allow users to configure their hosts to authenticate using SSH keys.

See [GitHub's instructions](https://help.github.com/articles/generating-ssh-keys/) or
[Bitbucket's instructions](https://confluence.atlassian.com/bitbucket/set-up-ssh-for-git-728138079.html)
for how to set this up.

**Note: For Linux distribution installations, Sourcegraph runs using the `sourcegraph`.
If setting up SSH mirroring on a Linux installation, be sure to generate an SSH keypair from the
`sourcegraph` user and add the public key with your git hosting service.**

Determine the SSH clone URL for your repository. It typically has the form
`git@bitbucket.org:userorganization/test_repository.git` and can be found on
your GitHub or Bitbucket repository page.

Check that git commands succeed with this URL. **On Linux distributions,
make sure to run the command as the `sourcegraph` user, otherwise the `sourcegraph` known_hosts
file will not contain your git server!**:

```
git ls-remote git@bitbucket.org/userorganization/test_repository.git
```

If that succeeds, you can now configure Sourcegraph to mirror your private repository
using the `src` CLI:

```
src repo create -m --clone-url=git@bitbucket.org/userorganization/test_repository.git <repo-name>
```

### 2. Authenticating over HTTPS

Alternatively, you can mirror a private repository with an HTTPS git clone URL.
It typically has the form `https://bitbucket.org/userorganization/test_repository`.

Add your username and password to the URL like this:
`https://user:password@bitbucket.org/userorganization/test_repository.git`.

Check that git commands succeed with this URL:

```
git ls-remote https://user:password@bitbucket.org/userorganization/test_repository
```

If that succeeds, you can now configure Sourcegraph to mirror your private repository
using the `src` CLI:

```
src repo create -m --clone-url=https://user:password@bitbucket.org/userorganization/test_repository.git <repo-name>
```

## 3. Using GitHub access tokens

Sourcegraph supports importing private repositories from GitHub.com and GitHub Enterprise
via GitHub personal access tokens.

Follow instructions for [GitHub.com]({{< relref "integrations/GitHub.md" >}}) or
[GitHub Enterprise]({{< relref "integrations/GitHub_Enterprise.md" >}}).

# Mirroring *to* externally hosted repositories

In addition to reading **from** an external host, Sourcegraph can automatically push
a repository **to** an external host.

## 1. Ensure you can push changes

First, get shell access to your Sourcegraph server (`src.example.com`) and ensure that you can push changes to `github.com/my/repo` _without specifying a username/password on the command line_ (you can do this by configuring your SSH keys as you would normally).

If you skip this step, Sourcegraph won't be able to push to the remote repository.

## 2. Add config option

Now you'll need to either use the `src serve --fs.git-repo-mirror` CLI flag or add this to your configuration file:

```
[serve]
fs.git-repo-mirror = <LocalRepoURI>:<GitRemoteURL>
```

More than one repo can be mirrored by simply adding a comma. For example:

```
[serve]
fs.git-repo-mirror = my/repo:git@github.com:my/repo,other/hacks:git@github.com:me/hacks
```

Would cause any pushes to `src.example.com/my/repo` to be mirrored to `github.com/my/repo`.
Likewise, any changes to `src.example.com/other/hacks` will be mirrored to `github.com:me/hacks`.

## 3. Restart Sourcegraph

After updating your config, restart the `src` process.

## 4. Test it out!

Now that everything is configured, simply push any change to your Sourcegraph repository and watch as it is mirrored to the remote git URL you specified!
