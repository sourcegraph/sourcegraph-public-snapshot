+++
title = "Git over SSH"
+++

Sourcegraph supports HTTP and SSH transfer protocols for git operations.

To use the SSH transfer protocol, first add your public key to Sourcegraph:

```
src --endpoint=https://src.mycompany.com login
src users keys add <public-key-file>
```

Then, do some git operations:

```
git remote add src `ssh://git@src.mycompany.com:3022/path/to/repo`
git push src master
```

The git transport will be done over SSH using your keypair.

If you'd like to change the ssh port of your Sourcegraph server, set
the `--ssh-addr` flag or update your `config.ini` file:

```
[serve]
SSHAddr = 22
```
