# Cody App release pipeline

The Cody App release pipeline utilizes buildkite to build and bundle Cody App for all current [supported platforms](#supported-platforms). The primary definition of the buildkite pipeline can be found at `.buildkite/pipeline.app.yml` in the Sourcegraph mono repo.

## Branches that trigger the release pipeline

As soon as a commit is pushed on the following branches the release pipeline will start. Note that only one build can be in progress at one time, and pushing a new commit will cancel all previous builds.

- `app-release/stable`: Once a stable release is ready commits should be pushed to this branch ex. `git push HEAD:app-release/stable` will push the current commit on your current branch (which should be main) to `app-release/stable`.
- `app-release/debug`: To debug the release pipeline, commits can be pushed to this branch ex. `git push HEAD:app-release/debug` will push the current commit on your current branch (which should be main) to `app-release/debug`.

For now the only difference between the two branches is how the release will be named:

- Releases from the `app-release/stable` branch will use the build number in their version number.
- Releases from the `app-release/debug` branch will use `debug` in the version number.

## Supported platforms

Cody App currently supports the following platforms:

- `x86_64-linux`
- `x86_64-darwin` (also known as Intel Mac)
- `aarch64-darwin` (also known as Apple Silicon)

## Broad overview of the pipeline

The pipeline is broken up into 3 stages:

1. Build the Sourcegraph Backend.
2. Bundle the Sourcegraph Backend using Tauri for the various supported platforms as well as perform code signing.
3. Publish a **draft** release on GitHub.

The above stages are accomplished by utilizing hosts across two cloud providers namely GCP and AWS.

- For GCP we execute any step that doesn't require tooling specific APIs or tooling.
  - Compiling and bundling of Cody App for Linux platform.
    - We **only** utilize Bazel to compile the Sourcegraph Backend.
  - Performing the GitHub release.
- We use AWS to host a MacOS host, which we utilize for all Apple specific tooling and processes.
  - Code Signing.
  - Compiling and bundling of Cody App for Intel Mac and Apple Silicon.
    - We use Bazel to compile the Sourcegraph Backend for `aarch64-darwin`.
    - We use Go to compile the Sourcegraph Backend for `x86_64-darwin`.

### Building of Sourcegraph Backend

The script that builds the backend is located at [`dev/app/build-backend.sh`](https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph@4cb14a729d2bcd86b47c6ee65f6fe7e34d2ff782/-/blob/enterprise/dev/app/build-backend.sh) and compiled binaries will be put under the directory `.bin` with the following naming convention `sourcegraph-backend-{platform}`. The script by default will use Bazel to compile the Sourcegraph backend and will automatically detect the platform by using the script [`enterprise/dev/app/detect_platform.sh`](https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph@4cb14a729d2bcd86b47c6ee65f6fe7e34d2ff782/-/blob/enterprise/dev/app/detect-platform.sh). For cross compilation and especially for the `x86_64-darwin` platform, one has to provide the following environment variables to make the script do the cross compilation:

- `PLATFORM=x86_64-apple-darwin`
- `CROSS_COMPILE_X86_64_MACOS=1`

### Bundling with Tauri

Tauri is used to provide a cross platform native shell for Sourcegrah App. The pipeline executes the script [`tauri-build.sh`](https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph@4cb14a729d2bcd86b47c6ee65f6fe7e34d2ff782/-/blob/enterprise/dev/app/tauri-build.sh) which starts the bundling process. Without going into too much detail of the script it performs the following functions:

- Downloads the various platform specific `sourcegraph-backend` binaries from the previous build steps
- On MacOS the `sourcegraph-backend` is sent to Apple for precode signing
- Creates the platform specific bundle
  - For Linux the bundles are `.deb` and `.AppImage`
  - For Darwin the bundles are `.dmg`, `aarch64.app.tar.gz` and `x64.app.tar.gz`.
- Tauri signs the following bundles and generates `.sig` files with the signatures for the following bundles:
  - `.AppImage`
    - Note the `.deb` bundle contains the `.AppImage` bundle.
  - `.aarch64.app.tar.gz` and `.x64.app.tar.gz`
    - The `.dmg` bundle contains the `.app.tar.gz` bundle.
- The bundles along with their signatures are uploaded

### Creating the GitHub release

Creation of the GitHub release is driven by the [`create-github-release.sh`](https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph@4cb14a729d2bcd86b47c6ee65f6fe7e34d2ff782/-/blob/enterprise/dev/app/create-github-release.sh) script, and crucially also generates the `app.update.manifest`. The script downloads all the artefacts from the previous steps and then generates a release using the GitHub cli. Note that the release will have the following values:

- Tag `app-v{VERSION}`
- Draft
- Prerelease

At this point in time no release notes are generated but that is a planned feature to add.

Finally, once the release has been created the script will generate an `app.update.manifest` by executing [`create-update-manifest.sh`](https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph@4cb14a729d2bcd86b47c6ee65f6fe7e34d2ff782/-/blob/enterprise/dev/app/create-update-manifest.sh) which has the following format:

```
{
  "version": "2023.5.30+1306.85e549168c",
  "notes": "",
  "pub_date": "2023-05-30T22:33:56Z",
  "platforms": {
    "aarch64-darwin": {
      "signature": "dW50cnVzdGVkIGNvbW1lbnQ6IHNpZ25hdHVyZSBmcm9tIHRhdXJpIHNlY3JldCBrZXkKUlVRMXp3Y3ZEa1JXajgyZk1CYWZkQjFrWjdzU0ZNT0twQ3ZESm1YMDVhZ3U5MTIycGFFakUwUElKOUt2N0JIZXhCZE9NaVgwSHhrYjZFcU42TDBEaGVlcW1QRXRnNzNuMFFJPQp0cnVzdGVkIGNvbW1lbnQ6IHRpbWVzdGFtcDoxNjg1NDg1NjMxCWZpbGU6U291cmNlZ3JhcGguYXBwLnRhci5negp1RUZLT1B4c0lJcWI3YlI5MzlXM0lKSU5McGJsQ2RNM0JsVXVUenhzTjJRSHpEWGpOMWpyVkc2ZkYwam1hYi9maSs4MnBkYWliK09GRml4ZXUwTVdBQT09Cg==",
      "url": "https://github.com/sourcegraph/sourcegraph/releases/download/untagged-app-v2023.5.30%2B1306.85e549168c/Sourcegraph.2023.5.30%2B1306.85e549168c.aarch64.app.tar.gz"
    },
    "x86_64-darwin": {
      "signature": "dW50cnVzdGVkIGNvbW1lbnQ6IHNpZ25hdHVyZSBmcm9tIHRhdXJpIHNlY3JldCBrZXkKUlVRMXp3Y3ZEa1JXanpyWXJqd3dUcnZVczRES052REZybjdkaWNHYjY1alVnQjYxaERyOTZURHdnMVBrMlVWUmtaRFFIYnpBREN2NkxVNFFKM3B4V3pHdFlTZDhWYnliR3dVPQp0cnVzdGVkIGNvbW1lbnQ6IHRpbWVzdGFtcDoxNjg1NDg1OTY3CWZpbGU6U291cmNlZ3JhcGguYXBwLnRhci5negpFdGg3cEVIVk90L1owUHhJQ0hBY0o1UkgwZllJdHRFNnJ6Ly9hSzQ0WURkNU5zcDdaK2RCcWhsSjRNVWF4NXpIeURsSWVwYnZKaElGK0RPc1cwMktEQT09Cg==",
      "url": "https://github.com/sourcegraph/sourcegraph/releases/download/untagged-app-v2023.5.30%2B1306.85e549168c/Sourcegraph.2023.5.30%2B1306.85e549168c.x86_64.app.tar.gz"
    },
    "x86_64-linux": {
      "signature": "dW50cnVzdGVkIGNvbW1lbnQ6IHNpZ25hdHVyZSBmcm9tIHRhdXJpIHNlY3JldCBrZXkKUlVRMXp3Y3ZEa1JXandyRWNReDc1TGZqTEFJaUxYcEtjWWFHSEpJZndkTnkrQVlKWSt2SHRXRGcxME5LbHhzenRYVktSdHE3YnRYQk90TjFIeXBjcVIweXBqWENuL3cyVHdJPQp0cnVzdGVkIGNvbW1lbnQ6IHRpbWVzdGFtcDoxNjg1NDg1OTQxCWZpbGU6c291cmNlZ3JhcGhfMjAyMy41LjMwKzEzMDYuODVlNTQ5MTY4Y19hbWQ2NC5BcHBJbWFnZS50YXIuZ3oKVkpSa3YvYStUTEVJQytsK21IWUtXWXFXOVp0Mk9FQUVuUTB4YUk1N0w3V1dCV0p5UzJCRkM0bjQ5NGdIQUJCaGF3VUN4UHlhZEl1UnVzN3YwYXM3RHc9PQo=",
      "url": "https://github.com/sourcegraph/sourcegraph/releases/download/untagged-app-v2023.5.30%2B1306.85e549168c/sourcegraph_2023.5.30%2B1306.85e549168c_amd64.AppImage.tar.gz"
    }
  }
}
```

The manifest will be available as an artefact on the `Create GitHub release` step of the pipeline. The manifest is used by Sourcegraph.com to tell Cody App clients checking in whether an update is available. Therefore, this manifest has to be uploaded to the following buckets depending on the use case:

- Production
  - Bucket [`sourcegraph-app`](https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/cmd/frontend/internal/app/updatecheck/app_update_checker.go?L25): This bucket is available in the `sourcegraph-dev` project, and the manifest should have to following name [`app.update.prod.manifest.json`](https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/cmd/frontend/internal/app/updatecheck/app_update_checker.go?L31).
- Development
  - Bucket [`sourcegraph-app-dev`](https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/cmd/frontend/internal/app/updatecheck/app_update_checker.go?L28): This is bucket is available in the `sourcegraph-ci` project, and the manifest should have the following name [`app.update.prod.manifest.json`](https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/cmd/frontend/internal/app/updatecheck/app_update_checker.go?L31).

__Note__: if you're wondering why the bucket names are different even though we use two different projects, it is because GCP requires bucket names to be **globally** unique.

### Automatically updating the manifest

`sg` can be used to automatically update the manifest from the latest build on `app-release/stable` and corresponding latest release on GitHub. Updating the manifest will set the new current version for download links as well as prompt users to upgrade the next time the App checks for  updates. To run the update-manifest command request the `Cody App Release Bucket access` bundle in Entitle.

Below is an example command:

```
sg app update-manifest --bucket sourcegraph-app
```

If one first wants to check the manifest the `--no-upload` flag can be passed which will print out the manifest that will get uploaded
```
sg app update-manifest --bucket sourcegraph-app --no-upload
```

Additionally, the manifest will be updated with the release notes from the GitHub release. By doing this the release notes are presented to the user when they are notified of an available update.
