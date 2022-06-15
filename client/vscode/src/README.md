# Extension Source Code

This folder contains the source code for the extension.

## Contexts

As detailed in our CONTRIBUTING.md file, this extension runs code in the following execution contexts:

1. Core App

- The main app where all the commands, webviews, file system, and all other components are put together and registered as a single extension
  - Runs code with Node.js on Desktop clients (for example, VS Code on Desktop) as a regular extension, or Web Worker on Web clients (for example, github.dev on browser) as a web extension
  - Web extension is run in the web extension host in a Browser WebWorker environment, and do not have access to Node.js globals and libraries at runtime

2. Sidebars

- This is the UI for all the sidebars used in the core app
  - Uses the VS Code Webview API to render the webview views in search sidebar, search history sidebar, auth sidebar, and help and feedback sidebar

3. Search Panel

- This is the UI for the extension search homepage and to display search results
  - Uses the VS Code Webview API to render the webview panel for the Sourcegraph search homepage as a distinct editor tab

4. Extension Host

- It processes the code-intel data
  - Uses Web Worker

5. Custom File System

- This is a custom file system implemented to display search result in a virtual workspace
  - Displays file-tree for the selected repository using the VS Code File System API
  - Opens result files that do not exist locally as virtual documents using the VS Code text document content provider AP

## List of Files and Folders

See the File Structure section in our CONTRIBUTING.md file to learn more about the folders and files in this directory.
