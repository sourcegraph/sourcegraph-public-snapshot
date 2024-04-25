# Appliance development

## Running tests

To (re)generate golden fixtures, pass the following argument to `go test`:

```
go test -args appliance-update-golden-files
```

In order to run `go test` (with or without arguments), you must have
`setup-envtest` available:

```
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
```

This is not a requirement for the Bazel environment (`bazel test
:appliance_test` in this module).
