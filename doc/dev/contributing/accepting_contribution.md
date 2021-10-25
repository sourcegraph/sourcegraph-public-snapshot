# How to accept an external contribution

## CLA-bot

1. Check if a contributor signed the CLA [here](https://docs.google.com/spreadsheets/d/1_iBZh9PJi-05vTnlQ3GVeeRe8H3Wq1_FZ49aYrsHGLQ/edit?usp=sharing). All fields should be filled with valid data to proceed with the pull request.
2. If the CLA is signed — update the CLA-bot configuration [here](https://github.com/sourcegraph/clabot-config/edit/main/.clabot) by adding a contributor name to the `contributors` field, preserving the alphabetical order.
3. Comment on the pull request: `@cla-bot check`.
4. The `verification/cla-signed` workflow should become green. 🎉

## Buildkite

To request a Buildkite build for a pull request from a fork, check out the branch and use `sg ci build` after reviewing the code.

It is possible to check out the branch without having to re-clone the forked repo by running `git fetch git@github.com:theirusername/sourcegraph theirbranch:theirbranch`
Make sure that the created branch name exactly matches their branch name, otherwise Buildkite will not match the created build to their branch.
