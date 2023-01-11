# Sourcegraph App

- To develop on App, run `sg start enterprise-single-program`.
- To build a binary to use for yourself for development purposes, run the above command and then grab the `.bin/sourcegraph` file.
- To build and release a snapshot for other people to use, push a commit to the special `app/release-snapshot` branch (i.e., `git push -f origin HEAD:app/release-snapshot`). This runs the `../../dev/app/release.sh` script in CI to build and release a snapshot of Sourcegraph App.
