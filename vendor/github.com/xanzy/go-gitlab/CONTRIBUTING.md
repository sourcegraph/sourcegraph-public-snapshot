# How to Contribute

We want to make contributing to this project as easy as possible.

## Reporting Issues

If you have an issue, please report it on the [issue tracker](
https://github.com/xanzy/go-gitlab/issues).

When you are up for writing a PR to solve the issue you encountered, it's not
needed to first open a separate issue. In that case only opening a PR with a
description of the issue you are trying to solve is just fine.

## Contributing Code

Pull requests are always welcome. When in doubt if your contribution fits in with
the rest of the project, free to first open an issue to discuss your idea.

This is not needed when fixing a bug or adding an enhancement, as long as the
enhancement you are trying to add can be found in the public GitLab API docs as
this project only supports what is in the public API docs.

## Coding style

We try to follow the Go best practices, where it makes sense, and use [`gofumpt`](
https://github.com/mvdan/gofumpt) to format code in this project.

Before making a PR, please look at the rest this package and try to make sure
your contribution is consistent with the rest of the coding style.

New struct field or methods should be placed (as much as possible) in the same
order as the ordering used in the public API docs. The idea is that this makes it
easier to find things.

### Setting up your local development environment to Contribute to `go-gitlab`

1. [Fork](https://github.com/xanzy/go-gitlab/fork), then clone the repository.
    ```sh
    git clone https://github.com/<your-username>/go-gitlab.git
    # or via ssh
    git clone git@github.com:<your-username>/go-gitlab.git
    ```
1. Install dependencies:
    ```sh
    make setup
    ```
1. Make your changes on your feature branch
1. Run the tests and `gofumpt`
    ```sh
    make test && make fmt
    ```
1. Open up your pull request
