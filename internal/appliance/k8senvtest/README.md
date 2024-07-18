# k8senvtest

A wrapper package for sigs.k8s.io/controller-runtime/pkg/envtest. Has
compatibility with our bazel setup. Any package that makes us of this one should
add the following to the go_test directive in its BUILD.bazel:

```starlark
data = [
    "//dev/tools:kubebuilder-assets",
],
env = {
    "KUBEBUILDER_ASSET_PATHS": "$(rlocationpaths //dev/tools:kubebuilder-assets)",
},
```

And this should just work out of the box. See consumers of this package for
examples on how to use it, including safe teardown.
