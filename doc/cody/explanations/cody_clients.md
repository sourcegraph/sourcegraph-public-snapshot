# Cody clients, extensions, and plugins

There are two primary places where you can interface with Cody:

- The Sourcegraph product interface
- In your IDE, using Cody extensions & plugins

## Cody in the Sourcegraph UI

Cody chat can be accessed directly in the core Sourcegraph interface. This is available within full Sourcegraph Enterprise instances and in the Cody desktop app.

### Cody with Sourcegraph Enterprise

Cody is accessible via the web interface for Sourcegraph Enterprise in 2 ways:

- The "Cody AI" tab: Use the chat window to ask Cody questions or request Cody fix/analyze/write code snippets.
- In the Sourcegraph blog view: Use the "Ask Cody" sidebar to ask Cody questions. Or, double-click on a symbol in the blog view to get a list of Cody commands including explaining code, translating code language, or providing code smells.

### The Cody desktop application

The Cody app also provides a chat interface for Cody. Use this interface to request information from Cody or paste code snippets into the interface. You can also select repositories that you'd like Cody to fetch context from.

[See more information about the app](../../app/index.md).

## Cody IDE IDE extensions

Cody also integrates directly into your IDE using plugins and extensions. Using Cody in your IDE allows it to also make inline fixes and autocomplete suggestions directly in your code.

Cody currently supports:

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
- [Visual Studio Code](https://code.visualstudio.com/)

Coming soon:

- [Emacs](https://www.gnu.org/software/emacs/)
- [Neovim](https://neovim.io/)

See the full feature comparison for Cody IDE extensions: 

<table>
   <thead>
      <tr>
        <th>IDE</th>
        <th>Download Link</th>
        <th>Status</th>
        <th>Chat</th>
        <th>Inline Chat</th>
        <th>Recipes</th>
        <th>Single-line autocomplete</th>
        <th>Multi-line autocomplete</th>
        <th>Connect to the Cody app</th>
      </tr>
   </thead>
   <tbody>
      <tr>
        <td>VS Code</td>
        <td><a href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai">Download</a></td>
        <td><a href="#status-definitions" class="indexer-status">🟢</a></td>
        <td class="indexer-implemented-y">✓</td> <!-- Chat -->
        <td class="indexer-implemented-y">✓</td> <!-- Inline Chat -->
        <td class="indexer-implemented-y">✓</td> <!-- Recipes -->
        <td class="indexer-implemented-y">✓</td> <!-- Single-line autocomplate -->
        <td class="indexer-implemented-y">✓</td> <!-- Multi-line autocomplete -->
        <td class="indexer-implemented-y">✓</td> <!-- Connect to the Cody app -->
      </tr>
      <tr>
        <td>JetBrains</td>
        <td><a href="https://www.jetbrains.com/products/">Product list</a></td>
        <td><a href="#status-definitions" class="indexer-status">🟡</a></td>
        <td class="indexer-implemented-y">✓</td> <!-- Chat -->
        <td class="indexer-implemented-n">✗</td> <!-- Inline Chat -->
        <td class="indexer-implemented-y">✓</td> <!-- Recipes -->
        <td class="indexer-implemented-y">✓</td> <!-- Single-line autocomplate -->
        <td class="indexer-implemented-n">✗</td> <!-- Multi-line autocomplete -->
        <td class="indexer-implemented-y">✓</td> <!-- Connect to the Cody app -->
      </tr>
   </tbody>
</table>

### Status definitions
An IDE extension status is:

- 🟢 _Generally Available_: Available as a normal product feature, no major features are absent.
- 🟡 _Partially available_: Available, but not yet with full functionality. Still in active development.
- 🟠 _Beta_: Available in pre-release form on a limited basis. 
