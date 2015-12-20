+++
title = "Repositories"
description = "Configure how repositories behave"
+++

# Description and language

```
src repo update my/repo --lang go
src repo update my/repo --description 'This is my repository.'
```

# Default branch

Sourcegraph determines the default branch of a Git repository in the
same way that Git itself does. To change the default branch, you must
have a shell in the repository's directory on the server (*not your
local clone*).

```
cd $SGPATH/repos/myrepo
git symbolic-ref HEAD refs/heads/my-new-default-branch
```

# Viewing commits with srclib Code Intelligence data

Sourcegraph supports a special Git revision syntax, `REV^{srclib}`,
which means "the nearest ancestor to REV that has srclib Code
Intelligence data." For example, `master` might refer to a commit you
just pushed that has not been analyzed by srclib yet, but
`master^{srclib}` would refer to an older commit that has been
analyzed. You can use this special Git revision syntax in any
Sourcegraph URL.

To make this behavior the default, set the repository's default branch
to `master^{srclib}` (instead of just `master`) using the steps above
under *Default branch*.
