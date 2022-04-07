## Dev setup

Install the formatter (all editors).

```
npm install -g elm-format
```

### Editor setup

To get the best development experience, you really should take the time to just set up a couple of key things which is (a) format-on-save (b) diagnostics/annotations.
The below instructions are for VS Code. You'll have to configure other editors on your own (I give some hints below, but haven't configure tested on other editors).
If you want to try things out all noncommital but want to use Vim/Emacs, then I recommend just getting Vim/[Emacs](https://marketplace.visualstudio.com/items?itemName=lfs.vscode-emacs-friendly) bindings in VS Code because the instructions here will nsure you get a nice dev environment (format-on-save, diagnostics, etc).

**VS Code**

Install:

- [Elm extension](https://marketplace.visualstudio.com/items?itemName=Elmtooling.elm-ls-vscode) for VS Code.
- [Download FiraCode font](https://github.com/tonsky/FiraCode/releases/download/5.2/Fira_Code_v5.2.zip). Optional, but recommended because it's a nice font for rendering some syntactic operators like `|>`. To install, unzip. Then double click all the files in `ttf` and click "Install font".

Optional: If you chose to add the font, add to in VS Code in `settings.json`: Press `command+shift+P`, type `settings.json`, click `Open settings.json` (not the Default settings) one, then add this:

```json
    "editor.fontFamily": "Fira Code",
    "terminal.integrated.fontFamily": "Fira Code",
    "editor.fontLigatures": true,
```

Required: Add these editor settings:

```json
    "editor.formatOnSave": true,
    "elmLS.onlyUpdateDiagnosticsOnSave": true,
```

In total your `settings.json` should look like this:

```json
{
  // ... your other junk

  // optional font things
  "editor.fontFamily": "Fira Code",
  "terminal.integrated.fontFamily": "Fira Code",
  "editor.fontLigatures": true,

  // required editor things
  "editor.formatOnSave": true,
  "elmLS.onlyUpdateDiagnosticsOnSave": true
}
```

- Optional: if you want to feel like a superhero bind `Show Hover` to a good shortcut command. Click the gear in the bottom left of VS Code, go `Keyboard Shortcuts` search for `showHover` and edit the key binding for `editor.action.showHover`

- Optional: if you want to feel like a ninja too, search for `next problem` in `Keyboard Shortcuts` and bind that to something convenient.

**IntelliJ**

There's a plugin: https://plugins.jetbrains.com/plugin/10268-elm

**Vim**

Untested, but you want to install and configure Elm language server. You need to install at least:

```
npm install -g elm elm-test elm-format @elm-tooling/elm-language-server
```

And then look at `elm-vim`: https://github.com/elm-tooling/elm-vim and go from there.

**Emacs**

See elm-mode: https://github.com/jcollard/elm-mode

### Running

Let's make sure things worked. You need `npx` installed (probably it is, but you can check with `which npx`). Then, in this directory (`client/web/src/notebooks/blocks/compute/component`), run:

```
make standalone
```

This starts a live server that watches for changes and spits out errors if the
program doesn't compile. It can be useful to run this in a separate terminal
(or in a terminal in VS Code) because the errors are helpful when hacking
(especially to see the build succeeds). You'll get similar errors with
diagnostics too, so you don't strictly need to care about the terminal output
if your editor integration works.

Visit http://localhost:8000 in your browser. You should see a purple bar chart load. If not, DM @rijnard in slack.

Next, to make sure your development setup is working, open `src/Main.elm` and delete the text that says:

```
| RunCompute
```

Formatting should work on save (deletes an empty line if you left an empty
line), and your editor should now give a couple of diagnostics like:

```
  I cannot find a `RunCompute` variant:...
```

You'll also see similar error messages in the terminal where we started the live server.

Things work! Undo your change to fix the app again.
