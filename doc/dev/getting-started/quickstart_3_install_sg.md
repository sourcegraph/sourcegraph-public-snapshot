# Quickstart step 3: Install `sg`

At Sourcegraph we use [`sg`](https://github.com/sourcegraph/sourcegraph/tree/main/dev/sg), the Sourcegraph developer tool, to manage our local developer environment.

## Install `sg`

To install `sg`, do the following:

1. Navigate to the `sourcegraph` repository that you cloned in "[Cloning our repository](quickstart_2_clone_repository.md)"

    ```
    cd sourcegraph
    ```

2. Install `sg` by running the following command:

    ```
    ./dev/sg/install.sh
    ```

    If `sg` was not successfully installed into your `$PATH` you'll see instructions to make sure it is.

    > NOTE: Linux users need to make sure that `sg` is first in `$PATH`, before the system utility `sg`.

See the [`sg` README](https://github.com/sourcegraph/sourcegraph/tree/main/dev/sg) for more information or ask in the `#dev-experience` Slack channel.

<!-- omit in toc -->
[< Previous](quickstart_2_clone_repository.md) | [Next >](quickstart_4_start_docker.md)
