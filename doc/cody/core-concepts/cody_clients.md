# Cody clients, extensions, and plugins

There are three primary places where you can interface with Cody:

- Sourcegraph web app
- Cody desktop app
- In your IDE, using Cody extensions & plugins

## Cody in the web app

Cody chat can be accessed directly in the Sourcegraph web app. This is available within enterprise instances' web app and on sourcegraph.com.

### Cody desktop app

The Cody app also provides a chat interface for Cody. Use this interface to request information from Cody or paste code snippets into the interface. You can also select repositories that you'd like Cody to fetch context from.

[See more information about the app](./../overview/app/index.md).

## Cody IDE extensions

Cody also integrates directly into your IDE using plugins and extensions. Using Cody in your IDE allows it to also make inline fixes and autocomplete suggestions directly in your code.

Cody currently supports:

- [Visual Studio Code](https://code.visualstudio.com/)
- The JetBrains product family
  - [Android Studio](https://developer.android.com/studio)
  - [AppCode](https://www.jetbrains.com/objc/)
  - [CLion](https://www.jetbrains.com/clion/)
  - [GoLand](https://www.jetbrains.com/go/)
  - [IntelliJ IDEA](https://www.jetbrains.com/idea/) (Community and Ultimate editions)
  - [PhpStorm](https://www.jetbrains.com/phpstorm/)
  - [PyCharm](https://www.jetbrains.com/pycharm/) (Community and Professional editions)
  - [Rider](https://www.jetbrains.com/rider/)
  - [RubyMine](https://www.jetbrains.com/ruby/)
  - [WebStorm](https://www.jetbrains.com/webstorm/)

Coming soon:

- [Emacs](https://www.gnu.org/software/emacs/)
- [Neovim](https://neovim.io/)

See the full feature comparison for Cody IDE extensions:

This represents the _current availability_. Notes on future availability and feature parity is coming soon.

| Subject                      | Description                                             | VS Code | JetBrains IDEs <span class="badge badge-experimental">Experimental</span> | Web UI | App |
|-----------------------------|---------------------------------------------------------|:-:|:---------------:|:-:|:-:|
|Chat                         | ChatGPT-like chat functionality                         |✅|        ✅        |✅|✅|
|Code autocomplete                  | Start typing a line of code and Cody will suggest a completion|✅|        ✅        |✖️|✖️|
|Mutli-line code autocomplete      | Cody suggests multiple lines of a completion            |✅|       ✖️        |✖️|✖️|
|Inline chat                | Start chats and fix code from within the code editor                |✅|       ✖️        |✖️|✖️|
|Recipes                      | Predefined prompts |✅|        ✅        |✅|✖️|
|Multi-repo context          | Prompts can gather context from up to 10 repositories |✖️|       ✖️        |✅|✅|
|Context selection            | Specify repos you want the prompt to gather context from|✖️|       ✖️        |✅|✅|
|Connect to Cody app            | IDE uses Cody app as context|✅|        ✅        |✖️||

### Cody with Sourcegraph Enterprise

Cody is accessible via the web interface for Sourcegraph Enterprise in 2 ways:

- The "Cody" tab in the web app: Use the chat window to ask Cody questions or request Cody fix/analyze/write code snippets.
- In the Sourcegraph blob view: Use the "Ask Cody" sidebar to ask Cody questions. Or, double-click on a symbol in the blob view to get a list of Cody commands including explaining code, translating code language, or providing code smells.
