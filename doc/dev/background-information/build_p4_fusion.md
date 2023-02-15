# Building p4-fusion

In order to import Perforce depots into Sourcegraph we first convert them into git repositories. We use an open source tool called [p4-fusion](https://github.com/salesforce/p4-fusion):

> A fast Perforce depot to Git repository converter using the Helix Core C/C++ API as an attempt to mitigate the performance bottlenecks in git-p4.py.

[Building](https://github.com/salesforce/p4-fusion#build) p4-fusion can be a little tricky as it depends on some older libraries and also doesn't build on M1 Apple laptops. To get around this we use [nix](https://nixos.org).

## How to build

Below are the instructions for building p4-fusion locally, assuming you have the [Sourcegraph repository](https://github.com/sourcegraph/sourcegraph) checked out.

1. Follow [these instruction](https://nixos.org/download.html) to install Nix. (Tested with version 2.11.1)
2. Navigate to the root of your Sourcegraph directory
3. Run `nix build ".#p4-fusion" --verbose --extra-experimental-features nix-command --extra-experimental-features flakes`

If the build completes successfully you should have a `p4-fusion` binary in `./result/bin/p4-fusion` which you can copy somewhere in your `$PATH`
