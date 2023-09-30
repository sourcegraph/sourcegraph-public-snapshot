<style>

.th:first-child,
td:first-child {
   min-width: 200px;
}

.markdown-body table thead tr{
  border-top:0;
}

.markdown-body table th, .markdown-body table td {
    text-align: left;
    vertical-align: baseline;
    padding: 0.5714286em;
}

.markdown-body table tr:nth-child(2n) {
  background: unset;
}

.markdown-body table th, .markdown-body table td {
    border: none;
}
</style>

# Feature Parity Reference for Cody Clients

<p class="subtitle">This document compares Cody's features and capabilities across different clients.
</p>

Here's a feature parity matrix that compares the capabilities of Cody Clients across different platforms like VS Code, JetBrains, Neovim, Sourcegraph.com (Web UI), and Cody app.

## Authentication

| **Feature**               | **VS Code** | **JetBrains** | **Neovim** | **Web UI** | **App** |
|-------------------------|---------|-----------|--------|--------|-----|
| Between multiple accounts     |    ✓    |     ✓     |   ✓    |  N/A   |  x  |
| Auth through browsers     |    x    |     ✓     |   x    |  x   |  x  |

## Cody Agent

| **Feature**               | **VS Code** | **JetBrains** | **Neovim** | **Web UI** | **App** |
|-------------------------|---------|-----------|--------|--------|-----|
| Uses agent for Chat     |    ✓    |     ✓     |   ✓    |  N/A   |  x  |
| Uses agent for Recipes     |     ✓    |     ✓     |   ✓    |  N/A   |  x  |
| Uses agent for Autocomplete     |    ✓    |     ✓     |   ✓    |  N/A   |  x  |

## Cody app connection

| **Feature**               | **VS Code** | **JetBrains** | **Neovim** | **Web UI** | **App** |
|-------------------------|---------|-----------|--------|--------|-----|
| Get context as back-end     |    ✓    |     x     |   x    |  N/A   |  N/A  |
| Indexing a repo     |    ✓    |     x     |   x    |  N/A   |  ✓  |
| Show indexing progress     |    ✓    |     x     |   x    |  N/A   |  ✓  |

## Chat

| **Feature**               | **VS Code** | **JetBrains** | **Neovim** | **Web UI** | **App** |
|-------------------------|---------|-----------|--------|--------|-----|
| Talk to Cody     |    ✓    |     ✓     |   ✓    |  ✓   |  ✓  |
| Chat history     |    ✓    |     x     |   x    |  ✓   |  ✓  |
| Stop chat generating     |    ✓    |     x     |   x    |  ✓   |  ✓  |
| Edit sent messages     |    ✓    |     x     |   x    |  ✓   |  ✓  |
| Slash (`/`) commands     |    ✓    |     x     |   x    |  x   |  x  |
| Chat predictions     |    ✓    |     x     |   x    |  x   |  x  |
| Show context files     |    ✓    |     ✓     |   ✓    |  ✓   |  x  |
| Show context files     |    ✓    |     ✓     |   ✓    |  ✓   |  x  |
| Custom commands     |    ✓    |     x     |   x    |  x   |  x  |
| Clear chat history     |    ✓    |     ✓     |   x    |  ✓   |  ✓  |

## Code Autocomplete

| **Feature**               | **VS Code** | **JetBrains** | **Neovim** | **Web UI** | **App** |
|-------------------------|---------|-----------|--------|--------|-----|
| Single-line autocompletion     |    ✓    |     ✓     |   ✓    |  x   |  x  |
| Single-line, multi-part autocompletion     |    ✓    |     ✓     |   ✓    |  x   |  x  |
| Multi-line, inline autocompletion     |    ✓    |     ✓     |   ✓    |  x   |  x  |
| Enable/Disable by language     |    x    |     ✓     |   ✓    |  x   |  x  |
| Customize autocomplete colors     |    x    |     ✓     |   ✓    |  x   |  x  |
| Cycle through multiple completion suggestions     |    x    |     x     |   ✓    |  x   |  x  |

## Code Context

| **Feature**               | **VS Code** | **JetBrains** | **Neovim** | **Web UI** | **App** |
|-------------------------|---------|-----------|--------|--------|-----|
| Multi-repo context (10 repos)     |    x    |     x     |   x    |  ✓   |  ✓  |
| Repo selection for context     |    ✓    |     ✓     |   x    |  ✓   |  ✓  |
| Local repo context     |    ✓    |     x     |   x    |  x   |  x  |
| Embeddings     |    ✓    |     ✓     |   ✓    |  ✓   |  ✓  |
| Context UI     |    ✓    |     ✓     |   x    |  ✓   |  -  |

## Inline Chat

| **Feature**               | **VS Code** | **JetBrains** | **Neovim** | **Web UI** | **App** |
|-------------------------|---------|-----------|--------|--------|-----|
| Questions     |    ✓    |     x     |   ✓    |  x   |  x  |
| Fix-ups     |    ✓    |     x     |   ✓    |  x   |  x  |
| Touch     |    ✓    |     x     |   x    |  x   |  x  |

## Commands/Recipes

| **Feature**               | **VS Code** | **JetBrains** | **Neovim** | **Web UI** | **App** |
|-------------------------|---------|-----------|--------|--------|-----|
| Generate `docstring`     |    ✓    |     ✓     |   ✓    |  x   |  x  |
| Generate unit test     |    ✓    |     ✓     |   ✓    |  x   |  x  |
| Explain code     |    ✓    |     ✓     |   ✓    |  x   |  x  |
| Smell code     |    ✓    |     ✓     |   ✓    |  x   |  x  |
| Ask a question     |    ✓    |     ✓     |   ✓    |  x   |  x  |
| Reset chat     |    ✓    |     x     |   x    |  ✓   |  ✓  |
| Task instruction     |    x    |     x     |   ✓    |  x   |  x  |
| Restart Cody/Sourcegraph     |    x    |     x     |   ✓    |  x   |  x  |
| Toggle chat window     |    x    |     x     |   ✓    |  x   |  x  |
| Improve variable names     |    x    |     ✓     |   x    |  ✓   |  ✓  |
| Symf search     |    -    |     ✓     |   x    |  x   |  x  |
| Generate README     |x    |     ✓     |   x    |  x   |  x  |
| Commit message suggestion     |x    |     ✓     |   x    |  x   |  x  |

## Recipes

| **Feature**               | **VS Code** | **JetBrains** | **Neovim** | **Web UI** | **App** |
|-------------------------|---------|-----------|--------|--------|-----|
| Optimize Code     |    x    |     x     |   ✓    |  x   |  x  |
