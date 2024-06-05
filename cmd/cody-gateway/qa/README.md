# Basic E2E / smoke test suite for Cody Gateway

To run against local instance ($TOKEN is an Enterprise subscription's `slk_` token) - note that local config like `CODY_GATEWAY_ANTHROPIC_ACCESS_TOKEN` etc. must be set in `sg.config.overwrite.yaml` for this to work:

```sh
bazel test --runs_per_test=2 --test_output=all //cmd/cody-gateway/qa:qa_test --test_env=E2E_GATEWAY_ENDPOINT=http://localhost:9992 --test_env=E2E_GATEWAY_TOKEN=$TOKEN
```

To run against dev instance using dotcom user ($TOKEN is a user's `sgd_` token):

```sh
bazel test --runs_per_test=2 --test_output=all //cmd/cody-gateway/qa:qa_test --test_env=E2E_GATEWAY_ENDPOINT=https://cody-gateway.sgdev.org --test_env=E2E_GATEWAY_TOKEN=$TOKEN
```

To run against prod using a dotcom user ($TOKEN is a user's `sgd_` token):

```sh
bazel test --runs_per_test=2 --test_output=all //cmd/cody-gateway/qa:qa_test --test_env=E2E_GATEWAY_ENDPOINT=https://cody-gateway.sourcegraph.com --test_env=E2E_GATEWAY_TOKEN=$TOKEN
```

The `--runs_per_test=2` flag in snippet above ensures we don't hit a Bazel cache, and runs the test twice for good meausre.
