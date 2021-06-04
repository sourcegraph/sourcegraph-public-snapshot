# How to configure submodules

## Prerequisites

This document assumes that you are a [site administrator](https://docs.sourcegraph.com/admin).

## Steps
1. Make sure you have a `.gitmodules` file created in folder level
2. Setup the `.gitmodules` file following the format shown in our example below
3. We use the `.gitmodules` file to figure out where a submodule exists in the tree. When a user clicks on the module in the UI, they will be redirected to the repo for the module.

## Example

URL should be used instead of local/relative path when setting up the `.gitmodules` file for GitLab repos:

```
[submodule "test-repo-2"]
        path = test-repo-2
        url = https://github.com/test-repos/test-repo-2
```
