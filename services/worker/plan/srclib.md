# Running srclib toolchains in Sourcegraph

After modifying srclib toolchains, you'll need to rebuild the Docker
images that Sourcegraph uses to run the toolchains. See
`worker/dockerfiles/README.md` for instructions.

To test out Sourcegraph CI when you've made these changes, you can of
course create a repo and build it in the UI. But you can also run `src
check --debug` in any locally checked out repo (which does not even
need to exist on any Sourcegraph server). That runs the same CI
process but locally.

# Infer build and test configuration

By making the user, not srclib, responsible for building and installing deps, we
(1) greatly simplify srclib and (2) make it easier for users to customize
srclib.

For example, suppose your Java build needed to add special auth for Artifactory,
needed to run an install.sh script as well as `mvn package`, and needed Maven
4.9 not Maven 4.3. How would you specify all that in srclib? By just configuring
it using a standard CI system (Drone), not custom srclib stuff, it is easier and
simpler.

The principle is: srclib and srclib toolchains assume the build system has
already fetched deps, compiled, etc.

Implicit srclib configuration is when there is no `.sg-drone.yml` and Sourcegraph
creates one based on the filenames (.java,. go, etc.) it sees. Explicit srclib
configuration is when you have a .sg-drone.yml with srclib steps in it; in that
case, no implicit configuration is performed (e.g., say you had some .py files
but didn't want to run srclib-python).

By default, Sourcegraphs adds up to one build step per language
* *Build* that tries to compile source code

Presence of this step depends on language (some may not have centralized
'build source code' entry point).

If build step defined in `.sg-drone.yml` or added implicitly refers to Docker image
built by Sourcegraph (identified by the presence of `srclib` substring in
image's name) it is assumed that indexing step was explicitly configured and
won't be added after the 'build' and 'test' steps. There is an exception to this:
**In Java-based projects, 'indexing' step is always added**.
