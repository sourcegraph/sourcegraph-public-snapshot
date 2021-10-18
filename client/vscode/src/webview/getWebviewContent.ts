import path from 'path'

import vscode from 'vscode'

export function getWebviewContent(extensionPath: string, webview: vscode.Webview, route: 'search'): string {
    const webviewPath = path.join(extensionPath, 'dist', 'webview')

    const scriptSource = webview.asWebviewUri(vscode.Uri.file(path.join(webviewPath, 'webview.js')))
    const cssModuleSource = webview.asWebviewUri(vscode.Uri.file(path.join(webviewPath, 'webview.css')))
    const styleSource = webview.asWebviewUri(vscode.Uri.file(path.join(webviewPath, 'style.css')))

    console.log({ route })

    // TODO: route through comlink, create only one application

    // TODO security
    return `<!DOCTYPE html>
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
}
