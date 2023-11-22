# Cody app Tauri shell

This contains all the Tauri code for the Cody App. Currently it consists of starting the `sourcegraph-backend` sidecar and then the Tauri frontend waits until it gets a message from the sidecar that the backend is up. Once the message has been received the Tauri frontend redirects the "web container" to http://localhost:3080.

## Required software

- Rust 1.73.0
- pnpm

## Getting started

1. Make sure you have all the required software. If not, you can run `sg setup` to get everything installed.

2. We have to make sure that all the client code has all its dependencies installed. To do that run `pnpm install` in the root of the repository.

3. Everything should now be ready! To do a full build without much thinking you can run `./dev/app/build-release.sh`, which will build the Tauri frontend shell, `sourcegraph-backend` and ultimately the Tauri App!

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
    ./cmd/sourcegraph
   ```

3. The previous command will put a binary in `.bin` named `sourcegraph-backend-aarch64-apple-darwin`. Take note of the name, which is `sourcegraph-backend`. This name has to match the sidecar name in the `tauri.conf.json` and in the Rust code!

### Tauri

Tauri has two modes that you can run.

- `Dev mode` which will run all of tauri but it does not embed the `sourcegraph-backend` instead it utilizes externally running sidecars and will directly open the app according to the `devPath` set in `tauri.conf.json`. The Cody App requires users to have a valid Sourcegraph.com account, before running the App ensure that an API Token for an account is provided in the site configuration at `app.dotcomAuthToken`. Dev mode utilizes `EXTSVC_CONFIG_ALLOW_EDITS=true` to allow local repositories to be added while App is being run. If you wish to run in dev mode with repos already added, they can be added to external services config like:

```
 {"url": "http://127.0.0.1:3434", "root": "PATH_TO_REPO(S)", "repos": ["src-serve-local"] }
```

To start the app in 'dev' mode you need to run the following command:

- `sg start app`
- `pnpm tauri build` this will build the complete Tauri bundle, which, if you're on mac, will be called `Cody.dmg`. The other noticeable difference is that is will run the `client/app-shell` code.

## Updating the icons

### Application icons

The application icon is defined in the `tauri.bundle.icon` field in `tauri.conf.json` and the files are located in `src-tauri/icons`. The application icon is one icon image, but is made available in multiple sizes and formats.

Do not create the sizes and formats manually for the application icon. Instead, use the `tauri icon` command.

```
cd src-tauri
tauri icon icon.png
```

The input file, `icon.png`, is expected to be a full-size PNG with transparency. It's recommended to be 1024x1024 in size.

The `tauri icon` command will generate all of the variants and sizes automatically, and place them in `src-tauri/icons`.

### System tray icon

The system tray icon is different from the application icon. It's placed in `icons/tray.png` and defined in `tauri.systemTray.iconPath`.

The tray icon is expected to be grayscale and will automatically be adjusted between light and dark modes. This behavior is enabled by the `iconAsTemplate` flag.

Note that the light and dark versions of the tray icon depend on the color of the desktop wallpaper, and not on the system UI light/dark theme.

## Where to get help or support

You can join us in Slack! We have a few channels that you might be interested in:

- Join #ask-app to ask any app related questions.
- Join #job-fair-app to get a constant stream of the progress that we're making on the app.
- Join #announce-app get app related news and announcements.
