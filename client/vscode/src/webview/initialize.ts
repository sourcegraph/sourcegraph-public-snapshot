import path from 'path'

import * as Comlink from 'comlink'
import { Observable } from 'rxjs'
import { filter, first } from 'rxjs/operators'
import vscode from 'vscode'

import {
    SourcegraphVSCodeExtensionAPI,
    SourcegraphVSCodeSearchSidebarAPI,
    SourcegraphVSCodeSearchWebviewAPI,
} from './contract'
import { createEndpointsForWebview } from './platform/extensionEndpoint'

interface SourcegraphWebviewConfig {
    extensionPath: string
    sourcegraphVSCodeExtensionAPI: SourcegraphVSCodeExtensionAPI
}

export async function initializeSearchPanelWebview({
    extensionPath,
    sourcegraphVSCodeExtensionAPI,
    initializedPanelIDs,
}: SourcegraphWebviewConfig & {
    initializedPanelIDs: Observable<string>
}): Promise<{
    sourcegraphVSCodeSearchWebviewAPI: Comlink.Remote<SourcegraphVSCodeSearchWebviewAPI>
    webviewPanel: vscode.WebviewPanel
}> {
    const panel = vscode.window.createWebviewPanel('sourcegraphSearch', 'Sourcegraph Search', vscode.ViewColumn.One, {
        enableScripts: true,
        retainContextWhenHidden: true, // TODO document. For UX
    })

    const webviewPath = path.join(extensionPath, 'dist', 'webview')

    const scriptSource = panel.webview.asWebviewUri(vscode.Uri.file(path.join(webviewPath, 'searchPanel.js')))
    const cssModuleSource = panel.webview.asWebviewUri(vscode.Uri.file(path.join(webviewPath, 'searchPanel.css')))
    const styleSource = panel.webview.asWebviewUri(vscode.Uri.file(path.join(webviewPath, 'style.css')))

    const { proxy, expose, panelId } = createEndpointsForWebview(panel)

    // Wait for the webview to initialize or else messages will be dropped
    const hasInitialized = initializedPanelIDs
        .pipe(
            filter(initializedPanelId => initializedPanelId === panelId),
            first()
        )
        .toPromise()

    // Get a proxy for the Sourcegraph Webview API to communicate with the Webview.
    const sourcegraphVSCodeSearchWebviewAPI = Comlink.wrap<SourcegraphVSCodeSearchWebviewAPI>(proxy)

    // Expose the Sourcegraph VS Code Extension API to the Webview.
    Comlink.expose(sourcegraphVSCodeExtensionAPI, expose)

    // TODO(tj): SECURITY!!! temporary script-src unsafe-eval for development mode
    panel.webview.html = `<!DOCTYPE html>
    <html lang="en" data-panel-id="${panelId}">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src vscode-resource: https:; script-src 'unsafe-eval' vscode-resource:;style-src vscode-resource: 'unsafe-inline' http: https: data:; connect-src 'self' http: https:;">
        <title>Sourcegraph Search</title>
        <link rel="stylesheet" href="${styleSource.toString()}" />
        <link rel="stylesheet" href="${cssModuleSource.toString()}" />
    </head>
    <body>
        <div id="root" />
        <script src="${scriptSource.toString()}"></script>
    </body>
    </html>`

    await hasInitialized

    return {
        sourcegraphVSCodeSearchWebviewAPI,
        webviewPanel: panel,
    }
}

export function initializeSearchSidebarWebview({
    extensionPath,
    sourcegraphVSCodeExtensionAPI,
    webviewView,
}: SourcegraphWebviewConfig & {
    webviewView: vscode.WebviewView
}): {
    sourcegraphVSCodeSearchSidebarAPI: Comlink.Remote<SourcegraphVSCodeSearchSidebarAPI>
} {
    webviewView.webview.options = {
        enableScripts: true,
    }

    const webviewPath = path.join(extensionPath, 'dist', 'webview')

    const scriptSource = webviewView.webview.asWebviewUri(vscode.Uri.file(path.join(webviewPath, 'searchSidebar.js')))
    const cssModuleSource = webviewView.webview.asWebviewUri(
        vscode.Uri.file(path.join(webviewPath, 'searchSidebar.css'))
    )
    const styleSource = webviewView.webview.asWebviewUri(vscode.Uri.file(path.join(webviewPath, 'style.css')))

    const { proxy, expose, panelId } = createEndpointsForWebview(webviewView)

    // Get a proxy for the Sourcegraph Webview API to communicate with the Webview.
    const sourcegraphVSCodeSearchSidebarAPI = Comlink.wrap<SourcegraphVSCodeSearchSidebarAPI>(proxy)

    // Expose the Sourcegraph VS Code Extension API to the Webview.
    Comlink.expose(sourcegraphVSCodeExtensionAPI, expose)

    // TODO(tj): SECURITY!!! temporary script-src unsafe-eval for development mode
    webviewView.webview.html = `<!DOCTYPE html>
    <html lang="en" data-panel-id="${panelId}">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src vscode-resource: https:; script-src 'unsafe-eval' vscode-resource:;style-src vscode-resource: 'unsafe-inline' http: https: data:; connect-src 'self' http: https:;">
        <title>Sourcegraph Search Sidebar</title>
        <link rel="stylesheet" href="${styleSource.toString()}" />
        <link rel="stylesheet" href="${cssModuleSource.toString()}" />
    </head>
    <body>
        <div id="root" />
        <script src="${scriptSource.toString()}"></script>
    </body>
    </html>`

    return {
        sourcegraphVSCodeSearchSidebarAPI,
    }
}
