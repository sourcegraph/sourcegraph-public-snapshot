# GoLand

[GoLand](https://www.jetbrains.com/go/) is a popular IDE for developing Go projects from JetBrains. If you’re a GoLand user, you can find useful tips and tricks for setting up your GoLand to develop Sorucegraph.

This page isn’t an exhaustive list of general-purpose tips for using GoLand; it aims to remain Sourcegraph specific.

### Use .editorconfig

GoLand automatically picks up the `.editorconfig` file [committed at the root](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/.editorconfig) of our repository. You don't need to do anything, just make sure that you don't disable the bundled plugin.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/goland-editorconfig-plugin.png" class="lead-screenshot">

### Imports style

Configure your GoLand like this to respect our imports settings:

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/goland-imports.png" class="lead-screenshot">

### Fill paragraph for Go comments

Useful feature for making Go code comments look consistent across the codebase. We don't enforce the paragraph width at the moment, but we default to GoLand's default setting (80).

See it in action here:
https://www.jetbrains.com/go/guide/tips/fill-paragraph-for-go-comments/
