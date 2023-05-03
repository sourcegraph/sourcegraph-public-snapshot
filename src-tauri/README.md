# Sourcegraph app Tauri shell

This contains all the Tauri code for the Sourcegraph App. Currently it consists of starting the `sourcegraph-backend` sidecar and then the Tauri frontend waits until it gets a message from the sidecar that the backend is up. Once the message has been received the Tauri frontend redirects the "web container" to http://localhost:3080.

## Required software

- Rust 1.68.0
- pnpm

## Getting started

1. Make sure you have all the required software. If not, you can run `sg setup` to get everything installed.

2. We have to make sure that all the client code has all its dependencies installed. To do that run `pnpm install` in the root of the repository.

3. Everything should now be ready! To do a full build without much thinking you can run `./enterprise/dev/app/build-release.sh`, which will build the Tauri frontend shell, `sourcegraph-backend` and ultimately the Tauri App!

## How do I build the individual parts?

### Tauri frontend shell

You can find the Tauri frontend shell at `client/app-shell` and to build it you can either run `pnpm build-app-shell` from the repository root or in `client/app-shell` you can run `pnpm build`. Both commands will create a bundled js app in `client/app-shell/dist` along with an `index.html` which references the bundled js.

### Sourcegraph backend

The Sourcegraph backend, or "single binary" as it was previously known, contains all the Sourcegraph services but in one binary! To build it:

1. `pnpm build-web` - the Sourcegraph web app gets embedded into the frontend service binary, so we need to build it otherwise the frontend is going to have some old client web app (or nothing!)
2. ```
   go build \
    -o .bin/sourcegraph-backend-aarch64-apple-darwin \
    -tags dist \
    -ldflags '-X github.com/sourcegraph/sourcegraph/internal/conf/deploy.forceType=app' \
    ./enterprise/cmd/sourcegraph
   ```

3. The previous command will put a binary in `.bin` named `sourcegraph-backend-aarch64-apple-darwin`. Take note of the name, which is `sourcegraph-backend`. This name has to match the sidecar name in the `tauri.conf.json` and in the Rust code!

### Tauri

Tauri has two modes that you can run.

- `Dev mode` which will run all of tauri but it does not embed the `sourcegraph-backend` instead it utilizes externally running sidecars and will directly open the app according to the `devPath` set in `tauri.conf.json`. To run the app in 'dev' mode you need to run the following command:
  - `sg start app`
- `pnpm tauri build` this will build the complete Tauri bundle, which, if you're on mac, will be called `Sourcegrap App.dmg`. The other noticable difference is that is will run the `client/app-shell` code.

## Where to get help or support

You can join us in Slack! We have a few channels that you might be interested in:

- Join #ask-app to ask any app related questions.
- Join #job-fair-app to get a constant stream of the progress that we're making on the app.
- Join #announce-app get app related news and announcements.
