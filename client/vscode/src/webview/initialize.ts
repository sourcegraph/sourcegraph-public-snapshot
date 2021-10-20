import path from 'path'

import * as Comlink from 'comlink'
import vscode from 'vscode'

import { SourcegraphVSCodeExtensionAPI, SourcegraphVSCodeWebviewAPI } from './contract'
import { vscodeExtensionEndpoint } from './platform/extensionEndpoint'

interface SourcegraphWebviewConfig {
    extensionPath: string
    route: 'search'
    id: string
    title: string
    sourcegraphVSCodeExtensionAPI: SourcegraphVSCodeExtensionAPI
}

export function initializeWebview({
    extensionPath,
    route,
    id,
    title,
    sourcegraphVSCodeExtensionAPI,
}: SourcegraphWebviewConfig): void {
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
    Comlink.expose(sourcegraphVSCodeExtensionAPI, vscodeExtensionEndpoint(panel, 'extension'))

    // TODO security
    panel.webview.html = `<!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src vscode-resource: https:; script-src vscode-resource:;style-src vscode-resource: 'unsafe-inline' http: https: data:; connect-src 'self' http: https:;">
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
