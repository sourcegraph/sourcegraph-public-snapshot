# Contributing Guide

> NOTE: these instructions are for fixing or adding features to the `azopenai` module. To use the module refer to the readme for this package: [readme.md](https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/ai/azopenai/README.md).

This is a contributing guide for the `azopenai` package. For general contributing guidelines refer to [CONTRIBUTING.md](https://github.com/Azure/azure-sdk-for-go/blob/main/CONTRIBUTING.md).

The `azopenai` package can be used with either Azure OpenAI or OpenAI's public service. New features are added using our code generation process, specified using TypeSpec [TypeSpec](https://github.com/Microsoft/typespec), which details all the models and protocol methods for using OpenAI. 

### Prerequisites

For code fixes that do not require code generation:
- Go 1.18 (or greater)

For code generation:
- [NodeJS (use the latest LTS)](https://nodejs.org)
- [TypeSpec compiler](https://github.com/Microsoft/typespec#getting-started).
- [autorest](https://github.com/Azure/autorest/tree/main/packages/apps/autorest)
- [PowerShell Core](https://github.com/PowerShell/PowerShell#get-powershell)
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)

# Building

## Generating from TypeSpec

The `Client` is primarily generated from TypeSpec, with some handwritten code where we've changed the interface to match Azure naming conventions (for instance, we refer to Models as Deployments). Files that do not have `custom` (ex: `client.go`, `models.go`, `models_serde.go`, etc..) are generated.

Files that have `custom` in the name are handwritten (ex: `custom_client_audio.go`), while files that do not (ex: `client.go`, `models.go`, `models_serde.go`, etc..) are generated.

### Regeneration

The `testdata/tsp-location.yaml` specifies the specific revision (and repo) that we use to generate the client. This also makes it possible, if needed, to generate from branch commmits in [`Azure/azure-rest-api-specs`](https://github.com/Azure/azure-rest-api-specs).

**tsp.location.yaml**:
```yaml
# ie: https://github.com/Azure/azure-rest-api-specs/tree/1e243e2b0d0d006599dcb64f82fd92aecc1247be/specification/cognitiveservices/OpenAI.Inference
directory: specification/cognitiveservices/OpenAI.Inference
commit: 1e243e2b0d0d006599dcb64f82fd92aecc1247be
repo: Azure/azure-rest-api-specs
```
The generation process is all done as `go generate` commands in `build.go`. To regenerate the client run:

```
go generate ./...
```

Commit the generated changes as part of your pull request.

If the changes don't look quite right you can adjust the generated code using the `autorest.md` file.

# Testing

There are three kinds of tests for this package: unit tests, recorded tests and live tests.

## Unit and recorded tests

Unit tests and recorded tests do not require access to OpenAI to run and will run with any PR as a check-in gate. 

Recorded tests require the Azure SDK test proxy is running. See the instructions for [installing the test-proxy](https://github.com/Azure/azure-sdk-tools/blob/main/tools/test-proxy/Azure.Sdk.Tools.TestProxy/README.md#installation).

In one terminal window, start the test-proxy:

```bash
cd <root of the azopenai module>
test-proxy
```

In another terminal window:


To playback (ie: use recordings):
```bash
cd <root of the azopenai module>

export AZURE_RECORD_MODE=playback
go test -count 1 -v ./...
```

To re-record:
```bash
cd <root of the azopenai module>

export AZURE_RECORD_MODE=record
go test -count 1 -v ./...

# push the recording changes to the repo
test-proxy push -a assets.json

# commit our assets.json file now that it points
# to the new recordings.
git add assets.json
git commit -m "updated recordings"
git push
```

## Live tests

### Local development

Copy the `sample.env` file to `.env`, and fill out all the values. Each value is documented to give you a general idea of what's needed, but ultimately you'll need to work with the Azure OpenAI SDK team to figure out which services are used for which features. 

Once filled out, the tests will automatically load environment variables from the `.env`:

```bash
export AZURE_RECORD_MODE=live
go test -count 1 -v ./...
```

### Pull requests

Post a comment to your PR with this text:

```
/azp run go - azopenai
```

The build bot will post a comment indicating its started the pipeline and the checks will start showing up in the status for the PR as well.
