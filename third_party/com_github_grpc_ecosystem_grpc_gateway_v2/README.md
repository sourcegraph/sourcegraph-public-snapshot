## How to regenerate a patch

In the eventuality of having to regenerate this patch if we bump the package version, you can generate a new patch with the following:

```
    # Shallow clone based on version
    git clone --depth 1 --branch v2.16.0 https://github.com/grpc-ecosystem/grpc-gateway.git
    cd ./grpc-gateway

    # (Optional)
    # Remove bazel version if you need to run using different Bazel version
    rm .bazelversion

    # Add Gazelle directive to disable proto compilation
    echo '# gazelle:proto disable_global' >> BUILD

    # Run Gazelle update-repos command to update repositories.bzl with
    # disable_global flag
    bazel run gazelle -- update-repos \
        -from_file=go.mod \
        -to_macro=repositories.bzl%go_repositories \
        --build_file_proto_mode=disable_global

    # Remove BUILD.bazel file with conflicting import
    rm runtime/BUILD.bazel
    rm runtime/internal/examplepb/BUILD.bazel

    # Run Gazelle fix command to regenerate BUILD.bazel based on diasble_global
    bazel run gazelle -- fix

    # Create a patch file for two files that causes the build error:
    #   - `repositories.bzl`
    #   - `runtime/BUILD.bazel`
    #   - `runtime/internal/examplepb/BUILD.bazel`
    git diff -u repositories.bzl runtime/BUILD.bazel runtime/internal/examplepb/BUILD.bazel > ../grpc_gateway.patch
```
