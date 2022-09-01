# Open in Editor

With a click of a button, you can jump from any file in Sourcegraph to your preferred editor. It only takes a minute to
set up.

<img src="img/editors/appcode.svg" width="32" height="30" alt="AppCode logo" />
<img src="img/editors/clion.svg" width="32" height="30" alt="CLion logo" />
<img src="img/editors/goland.svg" width="32" height="30" alt="GoLand logo" />
<img src="img/editors/idea.svg" width="32" height="30" alt="IntelliJ logo" />
<img src="img/editors/phpstorm.svg" width="32" height="30" alt="PhpStorm logo" />
<img src="img/editors/pycharm.svg" width="32" height="30" alt="PyCharm logo" />
<img src="img/editors/rider.svg" width="32" height="30" alt="Rider logo" />
<img src="img/editors/rubymine.svg" width="32" height="30" alt="RubyMine logo" />
<img src="img/editors/webstorm.svg" width="32" height="30" alt="WebStorm logo" />
<img src="img/editors/sublime.svg" width="32" height="30" alt="Sublime logo" />
<img src="img/editors/vscode.svg" width="32" height="30" alt="Visual logo" />
<img src="img/editors/atom.svg" width="32" height="30" alt="Atom logo" />

## Get started

It only needs a minute to set up. If you haven’t used the feature, find this button on the right side of your screen:

<img src="img/open-in-editor.svg" width="20" height="20" alt="Open in Editor icon" style="margin:0 5px;" /> 

When you click it, a small popup appears where you’ll need to set two things: a project path and your editor.

- The **project path** is the folder where you store your projects. Enter any Linux, Windows, or macOS absolute path.
  - Example: If your project path is `/home/username/projects`, one of your repos
    is https://github.com/sourcegraph/sourcegraph/, and you want to open `README.md` in the root folder of the repo,
    then the feature will look for the file at `/home/username/project/sourcegraph/README.md`.
- Then choose your **preferred editor** from the list. We support VS Code, Atom, Sublime Text, and JetBrains IDEs (
  AppCode, CLion, GoLand, IntelliJ IDEA, PhpStorm, PyCharm, Rider, RubyMine, WebStorm, and all new IDEs that JetBrains
  releases). If you use a different editor, check out “Advanced config” below.

Once you click **Save**, the icon will you’ll be able to jump to your editor from any file in Sourcegraph.

## Advanced config

If you want to dig deeper, there are a bunch of optional settings to play with. Click your name at the top-right, select
Settings, and add the `"openInEditor": {}` key to your user setting JSON.

- **editorId:** a short name of an editor (“vscode”, “idea”, “sublime”, “pycharm”, etc.) or “custom”. This is what you
  set in the dropdown when you first set up the feature.
- **custom.urlPattern:** If you set editorId to “custom” then this must be set, too. Use the placeholders “%file”,
  “%line”, and “%col” to mark where the file path, line number, and column number must be insterted. Example URL pattern
  for IntelliJ IDEA: `idea://open?file=%file&line=%line&column=%col`
- **projectsPaths.default**: This is what you set in the form when you first set up the feature.
- **projectsPaths.linux**: Overwrites the default path on Linux. Handy if you use different environments.
- **projectsPaths.mac**: Overwrites the default path on macOS.
- **projectsPaths.windows**: Overwrites the default path on Windows.
- **replacements**: Key-value pairs. Each key will be replaced by the corresponding value in the final URL. Keys are
  regular expressions, values can contain backreferences ($1, $2, ...).
- **jetbrains.forceApi**: Forces using protocol handlers (like `idea://open?file=...`) or the built-in REST
  API (`http://localhost:63342/api/file...`). If omitted, **protocol handlers** are used if available, otherwise the
  built-in REST API is used.
- **vscode.isBasePathUNCPath**: Indicates that the given base path is a UNC (Universal Naming Convention) path.
- **vscode.useInsiders**: If set, files will open in VS Code Insiders rather than VS Code.
- **vscode.useSSH**: If set, files will open on a remote server via SSH. This requires vscode.remoteHostForSSH to be
  specified and VS Code extension “Remote Development by Microsoft” installed in your VS Code.
- **vscode.remoteHostForSSH**: The remote host as `USER@HOSTNAME`. This needs you to install the extension called
  “Remote Development by Microsoft” in your VS Code.

Just save your settings and enjoy being totally advanced! 😎
