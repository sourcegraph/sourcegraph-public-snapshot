# Contributing

First off, glad you're here and want to contribute! :heart:

## Getting Started

In order to work on this project, you'll need to install a few things:

1. A recent version of [Go](https://golang.org/doc/install)
1. [Docker](https://www.docker.com/) (only required to use `make dev/docker`, `make test/docker`, or release targets).
1. [unzip](http://infozip.sourceforge.net/)

Everything else that's needed will be installed as needed and put in `./bin`.

When writing tests, be sure that the package in the test file is suffixed with `_test`. Eg. `protoc_gen_doc_test`. This
ensures that you'll only be testing the public interface.

## Submitting a PR

Here are some general guidelines for making PRs for this repo.

1. [Fork this repo](https://github.com/pseudomuto/protoc-gen-doc/fork).
1. Make a branch off of master (`git checkout -b <your_branch_name>`).
1. Make focused commits with descriptive messages.
1. Add tests that fail without your code, and pass with it.
1. GoFmt your code! (see <https://blog.golang.org/go-fmt-your-code> to setup your editor to do this for you).
1. Be sure to run `make build/examples` so your changes are reflected in the example docs.
1. Running `make test/docker` should produce the same output as `make build/examples`.
1. **Ping someone on the PR** (Lots of people, including myself, won't get a notification unless pinged directly).

Every PR should have a well detailed summary of the changes being made and the reasoning behind them. Make sure to add
at least three sections.

### What is Changing?

Make sure you spell out in as much detail as necessary what will happen to which systems when your PR is merged, 
what are the expected changes.

### How is it Changing?

Include any relevant implementation details, mimize surprises for the reviewers in this section, if you had to take some 
unorthodox approaches (read hacks), explain why here.

### What Could Go Wrong?

How has this change been tested? In your opinion what is the risk, if any, of merging these changes.

#### Reviewers should:

1. Identify anything that the PR author may have missed from above.
1. Test the PR through whatever means necessary, including manually, to verify it is safe to be deployed.
1. Question everything. Never assume that something was tested or fully understood, always question and ask if there is
	 any uncertainty.
1. Before merging the PR make sure it has _**one**_ of the `Major release`, `Minor release`, or `Patch release` labels
	 applied to it (useful for the changelog and determining the version for the next release).

## Release Process

We follow [Semantic Versioning 2.0.0](http://semver.org/#semantic-versioning-200), and the fact that PRs are tagged with
the type of release makes determining the next version super simple.

Look through the new (since the last release) PRs that are included in this release to determine the new version. 

* If COUNT(labelled `Major release`) > 0, then it's a MAJOR version bump.
* If COUNT(labelled `Minor release`) > 0, then it's a MINOR version bump.
* PATCH version bump

### Now that we've got the version:

From an up-to-date master, do the following:

1. Run `make test/docker` to build the image and generate the examples. There should be no diff after this.
1. Update the version in `version.go` appropriately.
1. `git commit -am "Bump to version v<version>`.
1. `git tag -s v<version>`.
1. `git push origin master --tags`.

Once the tag is on GitHub, the release action will handle pushing to docker and creating a release in GitHub.
