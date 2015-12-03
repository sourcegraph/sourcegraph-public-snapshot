+++
title = "Git hosting troubleshooting"
linktitle = "Git hosting"
+++

To use Sourcegraph as your git repository host, simply
create a repository with the `src` command and point your git
operations to your Sourcegraph server, e.g.:
```
src --endpoint http://src.mycompany.com login
src repo create my/repo
```

Then, you may clone the empty repository:
```
git clone http://src.mycompany.com/my/repo
```

Or, you may push an existing repository from your local machine:
```
cd ~/path/to/my/repo
git remote add origin http://src.mycompany.com/my/repo
git push -u origin master
```

# Frequent issues / questions

## Login credentials

When prompted by git to provide your username/password,
use your Sourcegraph.com credentials.

### Invalid credentials

If using `osxkeychain` as your `credential.helper` you may have
cached login credentials; when your credentials change you may no
longer be able to use normal `git` operations.

To fix:
```
$ git credential-osxkeychain erase
host=src.mycompany.com
protocol=http
# [Press Return]
```

After this, you'll be prompted once again for your username/password when trying to
perform git operations to `http://src.mycompany.com`.

## Setting `origin`

If you are migrating your git repository from another host, you may
already have `origin` pointing to another domain and you cannot
run `git remote add origin ...` to create a new `origin`.

(If you do, you'll see `fatal: remote origin already exists`.)

To fix this, just run:
```
git remote set-url origin http://src.mycompany.com/my/repo
```
