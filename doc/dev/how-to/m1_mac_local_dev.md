# Workarounds for M1 Mac local development

## Rosetta

Docker [requires Rosetta](https://docs.docker.com/desktop/mac/apple-silicon/#system-requirements) to run `amd64` binaries. It should be installed by default, but if that wasn't the case, run `softwareupdate --install-rosetta`.

## Puppeteer
If you hit a Puppeteer error stating "The chromium binary is not available for arm64", you need to install Chromium for Puppeteer via Homebrew.
   ```
   brew install --no-quarantine chromium
   ```

   Check that Chromium opens correctly:
   ```
   open -a /Applications/Chromium.app
   ```

   If you hit an error above, try allowing the app under System Preferences > Security & Privacy > General Tab > Allow Anyways. If it was successful, exit Chromium and point Puppeteer to Chromium by adding the following to your shell configuration (e.g. `~/.zshenv`)
   ```
   export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true
   export PUPPETEER_EXECUTABLE_PATH=`which chromium`
   ```
   and updating your shell (e.g. `source ~/.zshenv`). 

(Based on https://linguinecode.com/post/how-to-fix-m1-mac-puppeteer-chromium-arm64-bug)

## Jaeger
[Get the Mac version of Jaeger](https://github.com/jhchabran/jaeger/releases/download/v1.28.1/jaeger-1.28.1-darwin-arm64.tar.gz), extract it, then 

```
# Set PROJECTS to where you're storing the sourcegraph repository.
export PROJECTS=~/work/
cd ~/Downloads
curl https://github.com/jhchabran/jaeger/releases/download/v1.28.1/jaeger-1.28.1-darwin-arm64.tar.gz -L | tar -xz
cp ~/Downloads/jaeger-1.28.1-darwin-arm64/jaeger-all-in-one $PROJECTS/sourcegraph/.bin/jaeger-all-in-one-1.18.1-darwin-arm64
``` 

(adjust `~/sourcegraph` to point to you local clone of `github.com/sourcegraph/sourcegraph`).

Did you bump into another issue and solve it locally? Consider updating this list! 🙇
