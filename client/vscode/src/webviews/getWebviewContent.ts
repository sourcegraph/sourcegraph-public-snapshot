import path from 'path'

import vscode from 'vscode'

export function getWebviewContent(extensionPath: string, webview: vscode.Webview, name: 'search'): string {
    const scriptPath = vscode.Uri.file(path.join(extensionPath, 'dist', 'webviews', `${name}.js`))
    const scriptSource = webview.asWebviewUri(scriptPath)

    // TODO security
    return `<!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <meta http-equiv="Content-Security-Policy" content="default-src self; img-src vscode-resource:; script-src vscode-resource: 'self' 'unsafe-inline'; style-src vscode-resource: 'self' 'unsafe-inline'; "/>
        <title>Sourcegraph Search</title>
    </head>
    <body>
        <div id="root" />
        <script src="${scriptSource.toString()}"></script>
    </body>
    </html>`
}
