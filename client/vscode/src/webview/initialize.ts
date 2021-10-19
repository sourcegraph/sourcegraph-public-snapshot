import path from 'path'

import * as Comlink from 'comlink'
import vscode from 'vscode'

import { SourcegraphVSCodeExtensionAPI, SourcegraphVSCodeWebviewAPI } from './contract'

interface SourcegraphWebviewConfig {
    extensionPath: string
    route: 'search'
    id: string
    title: string
}

export function initializeWebview({ extensionPath, route, id, title }: SourcegraphWebviewConfig): void {
    const panel = vscode.window.createWebviewPanel(id, title, vscode.ViewColumn.One, {
        enableScripts: true,
    })

    const webviewPath = path.join(extensionPath, 'dist', 'webview')

    const scriptSource = panel.webview.asWebviewUri(vscode.Uri.file(path.join(webviewPath, 'webview.js')))
    const cssModuleSource = panel.webview.asWebviewUri(vscode.Uri.file(path.join(webviewPath, 'webview.css')))
    const styleSource = panel.webview.asWebviewUri(vscode.Uri.file(path.join(webviewPath, 'style.css')))

    // Get a proxy for the Sourcegraph Webview API to communicate with the Webview.
    const sourcegraphVSCodeWebviewAPI = Comlink.wrap<SourcegraphVSCodeWebviewAPI>(
        vscodeExtensionEndpoint(panel, 'webview')
    )

    // Expose the Sourcegraph VS Code Extension API to the Webview.
    const sourcegraphVSCodeExtensionAPI: SourcegraphVSCodeExtensionAPI = {
        ping: () => 'pong!',
    }

    Comlink.expose(sourcegraphVSCodeExtensionAPI, vscodeExtensionEndpoint(panel, 'extension'))

    console.log('msgchan', globalThis.MessageChannel)

    // TODO security
    panel.webview.html = `<!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <meta http-equiv="Content-Security-Policy" content="default-src self; img-src vscode-resource:; script-src vscode-resource: 'self' 'unsafe-inline'; style-src vscode-resource: 'self' 'unsafe-inline'; "/>
        <title>Sourcegraph Search</title>
        <link rel="stylesheet" href="${styleSource.toString()}" />
        <link rel="stylesheet" href="${cssModuleSource.toString()}" />
    </head>
    <body>
        <div id="root" />
        <script src="${scriptSource.toString()}"></script>
    </body>
    </html>`

    sourcegraphVSCodeWebviewAPI.setRoute(route).then(
        () => {},
        () => {}
    )
}

/**
 * TODO explain
 */
function vscodeExtensionEndpoint(
    panel: vscode.WebviewPanel,
    connectionType: 'webview' | 'extension'
): Comlink.Endpoint {
    // TODO return endpoint and disposable fn to add to top level disposable?
    const listenerDisposables = new WeakMap<EventListenerOrEventListenerObject, vscode.Disposable>()

    return {
        postMessage: (message: any) => panel.webview.postMessage({ ...message, connectionType }),
        addEventListener: (type, listener) => {
            function onMessage(message: any): void {
                if (message?.connectionType === connectionType) {
                    // Comlink is listening for a message event, only uses the `data` property.
                    const messageEvent = {
                        data: message,
                    } as MessageEvent

                    return typeof listener === 'function' ? listener(messageEvent) : listener.handleEvent(messageEvent)
                }
            }

            const disposable = panel.webview.onDidReceiveMessage(onMessage)
            listenerDisposables.set(listener, disposable)
        },
        removeEventListener: (type, listener) => {
            listenerDisposables.delete(listener)
        },
    }
}
