+++
title = "Git Repo Mirroring"
+++

Sourcegraph works great as the system of record, but there are some situations in which you will want to mirror your repositories hosted on Sourcegraph to a Git repository hosted somewhere else. With configuration, Sourcegraph can automatically push changes to an external code host each time someone pushes changes to Sourcegraph.

# Configuration

Lets say you have a repository hosted on Sourcegraph (`src.example.com/my/repo`) and want to have this repository automatically mirrored on `github.com/my/repo`. Here's what you need to do:

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
