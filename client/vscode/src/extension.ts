// import { Endpoint } from 'comlink'
import vscode from 'vscode'

import { getWebviewContent } from './webviews/getWebviewContent'

export function activate(context: vscode.ExtensionContext): void {
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.test', () => {
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
            panel.webview.html = getWebviewContent(context.extensionPath, panel.webview, 'search')

            setInterval(() => {
                panel.webview.postMessage({ results: 'test' }).then(
                    () => {},
                    () => {}
                )
            }, 2000)
        })
    )
}

// TODO: Comlink VS Code adapter.
// function vscodeExtensionEndpoint(panel: vscode.WebviewPanel): Endpoint {
//     // TODO properly handle listener -> disposable assocation
//     let disposable: vscode.Disposable | undefined

//     return {
//         postMessage: (message: any) => panel.webview.postMessage(message),
//         addEventListener: (type, listener) => {
//             function onMessage(message: any): void {
//                 if (typeof listener === 'function') {
//                     return listener(message)
//                 }
//                 return listener.handleEvent(message)
//             }
//             disposable = panel.webview.onDidReceiveMessage(onMessage)
//         },
//         removeEventListener: () => {
//             disposable?.dispose()
//         },
//     }
// }

// interface WebviewVSCodeAPI {
//     postMessage: (message: any) => void
// }

// // Since webviews use window's addEventListener, we can probably use `windowEndpoint()` from comlink.
// function vscodeWebviewEndpoint(): Endpoint {}
