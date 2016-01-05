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
   https://src.sourcegraph.com/srclib.org/plugin/drone-srclib` repo's
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
