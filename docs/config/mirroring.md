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
