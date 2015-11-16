+++
title = "Git over SSH"
+++

Sourcegraph supports HTTP and SSH transfer protocols for git operations.

To use the SSH transfer protocol, first add your public key to Sourcegraph:

```
src --endpoint=https://src.mycompany.com login
src users keys add <public-key-file>
```

Then, configure a host which communicates with Sourcegraph over port 3022.
Open your `.ssh/config` (usually located at `~/.ssh/config`) and append the following
(replacing `<private-key-file>`):

```
Host src
  HostName https://src.mycompany.com
  Port 3022
  User root
  IdentityFile <private-key-file>
```

Now you can address the upstream repository as `git@src:path/to/repo.git`.
From your local reposotory root, try:

```
git remote add src `git@src:path/to/repo.git`
git push src master
```

The git transport will be done over SSH using your keypair.
