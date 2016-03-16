# Running srclib toolchains in Sourcegraph

A few changes need to be made to existing srclib toolchains to run
them in Sourcegraph as part of builds.

1. The toolchain needs to build against the srclib no-docker branch
   (soon to be merged into master), and in general, the toolchain's own
   Docker-awareness needs to be removed.

   This means:

    * IN_DOCKER_CONTAINER env var in code and in the Dockerfile
	* separate Makefile installation steps for inside Docker
	* eliminating now-unnecessary CMD/ENTRYPOINT directives from the Dockerfile
	* set the new toolchain Docker image tag to `srclib/srclib-LANG`, not `sqs1/srclib-LANG`
	* add Makefile targets `docker-image` that runs `docker build -t
      srclib/srclib-LANG .` and a `release` target that runs `docker
      push srclib/srclib-LANG` (these are by convention and are not
      enforced anywhere in code)
	* update the README.md if anything there is invalidated

	Also, this is now a good time to change the Dockerfiles to use
    standard base Docker images such as
    https://hub.docker.com/_/python/ (ideally the `-slim` variants,
    which take MUCH less disk space).
2. You need to create and push a new Docker image
   `srclib/drone-srclib-LANG`. Use the
   https://sourcegraph.com/sourcegraph/srclib.org/plugin/drone-srclib` repo's
   `dev` branch `toolchains/LANG` dir. Create a new dir based on the
   other dirs. This Docker image should be based on (`FROM`'d) on the
   `srclib/srclib-LANG` image you created earlier. It should download
   and install the `srclib` program and run `srclib config && srclib
   make` (the other `toolchains/LANG/**` files demonstrate this; just
   use those).
3. If the language is not already detected by `pkg/inventory` in the
   sourcegraph repo, add it there. NOTE: Soon we will move to using
   GitHub's linguist, which auto-detects hundreds of languages.
4. Add an automatic CI config gen for the language in
   `worker/plan/auto.go` (see how we're already doing it for Java, Go,
   JavaScript, etc.).
5. Add an automatic srclib config gen for the language in
   `worker/plan/srclib.go` (see how we're already doing it for Java, Go,
   JavaScript, etc.).

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

Implicit srclib configuration is when there is no `.drone.yml` and Sourcegraph
creates one based on the filenames (.java,. go, etc.) it sees. Explicit srclib
configuration is when you have a .drone.yml with srclib steps in it; in that
case, no implicit configuration is performed (e.g., say you had some .py files
but didn't want to run srclib-python).

By default, Sourcegraphs adds up to one build step per language
* *Build* that tries to compile source code

Presence of this step depends on language (some may not have centralized
'build source code' entry point).

If build step defined in `.drone.yml` or added implicitly refers to Docker image
built by Sourcegraph (identified by the presence of `srclib` substring in
image's name) it is assumed that indexing step was explicitly configured and
won't be added after the 'build' and 'test' steps. There is an exception to this:
**In Java-based projects, 'indexing' step is always added**.
