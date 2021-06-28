# `src` volume workspace base image

Sourcegraph `src` executes batch changes using either a bind or volume workspace. In the latter case (which is the default on macOS), this utility image is used to initialise the volume workspace within Docker, and then to extract the diff used when creating the changeset.

This image is based on Alpine, and adds the tools we need: curl, git, and unzip.

For more information, please refer to the [`src-cli` repository](https://github.com/sourcegraph/src-cli/tree/main/docker/batch-change-volume-workspace).

<!--
If you update this description, you _must_ also update the description at
https://hub.docker.com/r/sourcegraph/src-batch-change-volume-workspace — this
does not happen automatically!
-->
