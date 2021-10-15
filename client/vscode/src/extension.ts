import { Endpoint } from 'comlink'
import vscode from 'vscode'

import { areExtensionsSame } from '@sourcegraph/shared/src/extensions/extensions'

export function activate(context: vscode.ExtensionContext): void {
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.test', () => {
            const shouldDownloadExtensions = !areExtensionsSame([{ id: 'old' }], [{ id: 'new' }])
            console.log({ shouldDownloadExtensions })

            // Create and show panel
            const panel = vscode.window.createWebviewPanel(
                'sourcegraphSearch',
                'Sourcegraph Search',
                vscode.ViewColumn.One,
                {
                    enableScripts: true,
                }
            )

            // And set its HTML content
            panel.webview.html = getWebviewContent()

            setInterval(() => {
                panel.webview.postMessage({ results: 'test' }).then(
                    () => {},
                    () => {}
                )
            }, 2000)
        })
    )
}

function getWebviewContent(): string {
    return `<!DOCTYPE html>
  <html lang="en">
  <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>Sourcegraph Search</title>
  </head>
  <body>
      <input type="text" placeholder="search..." />

      <script>
            window.addEventListener('message', event => {
                console.log('msg', event.data)
            })
      </script>
  </body>
  </html>`
}

// TODO: Comlink VS Code adapter.
function vscodeExtensionEndpoint(panel: vscode.WebviewPanel): Endpoint {
    // TODO properly handle listener -> disposable assocation
    let disposable: vscode.Disposable | undefined

    return {
        postMessage: (message: any) => panel.webview.postMessage(message),
        addEventListener: (type, listener) => {
            function onMessage(message: any): void {
                if (typeof listener === 'function') {
                    return listener(message)
                }
                return listener.handleEvent(message)
            }
            disposable = panel.webview.onDidReceiveMessage(onMessage)
        },
        removeEventListener: () => {
            disposable?.dispose()
        },
    }
}

interface WebviewVSCodeAPI {
    postMessage: (message: any) => void
}

// Since webviews use window's addEventListener, we can probably use `windowEndpoint()` from comlink.
function vscodeWebviewEndpoint(): Endpoint {}
