import * as Comlink from 'comlink'
import { Observable } from 'rxjs'
import { filter, first } from 'rxjs/operators'
import vscode from 'vscode'

import { endpointSetting } from '../settings/endpointSetting'

import {
    SourcegraphVSCodeExtensionAPI,
    SourcegraphVSCodeExtensionHostAPI,
    SourcegraphVSCodeSearchSidebarAPI,
    SourcegraphVSCodeSearchWebviewAPI,
} from './contract'
import { createEndpointsForWebview } from './platform/extensionEndpoint'

interface SourcegraphWebviewConfig {
    extensionUri: vscode.Uri
    sourcegraphVSCodeExtensionAPI: SourcegraphVSCodeExtensionAPI
}

export async function initializeSearchPanelWebview({
    extensionUri,
    sourcegraphVSCodeExtensionAPI,
    initializedPanelIDs,
}: SourcegraphWebviewConfig & {
    initializedPanelIDs: Observable<string>
}): Promise<{
    sourcegraphVSCodeSearchWebviewAPI: Comlink.Remote<SourcegraphVSCodeSearchWebviewAPI>
    webviewPanel: vscode.WebviewPanel
}> {
    const panel = vscode.window.createWebviewPanel('sourcegraphSearch', 'Sourcegraph', vscode.ViewColumn.One, {
        enableScripts: true,
        retainContextWhenHidden: true, // TODO document. For UX
        localResourceRoots: [vscode.Uri.joinPath(extensionUri, 'dist', 'webview')],
    })

    const webviewPath = vscode.Uri.joinPath(extensionUri, 'dist', 'webview')

    const scriptSource = panel.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'searchPanel.js'))
    const cssModuleSource = panel.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'searchPanel.css'))
    const styleSource = panel.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'style.css'))

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

    // Specific scripts to run using nonce
    const nonce = getNonce()

    panel.iconPath = vscode.Uri.joinPath(extensionUri, 'dist', 'webview', 'logo.svg')

    // Apply Content-Security-Policy
    // panel.webview.cspSource comes from the webview object
    panel.webview.html = `<!DOCTYPE html>
    <html lang="en" data-panel-id="${panelId}">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src data: vscode-resource: vscode-webview: https:; script-src 'nonce-${nonce}' vscode-webview:; style-src data: ${
        panel.webview.cspSource
    } vscode-resource: vscode-webview: 'unsafe-inline' http: https: data:; connect-src 'self' vscode-webview: http: https:; frame-src https:; font-src: https: vscode-resource: vscode-webview:;">
        <title>Sourcegraph Search</title>
        <link rel="stylesheet" href="${styleSource.toString()}" />
        <link rel="stylesheet" href="${cssModuleSource.toString()}" />
    </head>
    <body>
        <div id="root" />
        <script nonce="${nonce}" src="${scriptSource.toString()}"></script>
    </body>
    </html>`

    await hasInitialized

    return {
        sourcegraphVSCodeSearchWebviewAPI,
        webviewPanel: panel,
    }
}

export function initializeSearchSidebarWebview({
    extensionUri,
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

    const webviewPath = vscode.Uri.joinPath(extensionUri, 'dist', 'webview')

    const scriptSource = webviewView.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'searchSidebar.js'))

    const cssModuleSource = webviewView.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'searchSidebar.css'))

    const styleSource = webviewView.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'style.css'))

    const { proxy, expose, panelId } = createEndpointsForWebview(webviewView)

    // Get a proxy for the Sourcegraph Webview API to communicate with the Webview.
    const sourcegraphVSCodeSearchSidebarAPI = Comlink.wrap<SourcegraphVSCodeSearchSidebarAPI>(proxy)

    // Expose the Sourcegraph VS Code Extension API to the Webview.
    Comlink.expose(sourcegraphVSCodeExtensionAPI, expose)

    // Specific scripts to run using nonce
    const nonce = getNonce()

    // Apply Content-Security-Policy
    // panel.webview.cspSource comes from the webview object
    webviewView.webview.html = `<!DOCTYPE html>
    <html lang="en" data-panel-id="${panelId}">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src data: vscode-webview: vscode-resource: https:; script-src 'nonce-${nonce}' vscode-webview:; style-src data: ${
        webviewView.webview.cspSource
    } vscode-resource: http: https: data:; connect-src 'self' http: https:; font-src: https: vscode-resource: vscode-webview:;">
        <title>Sourcegraph Search</title>
        <link rel="stylesheet" href="${styleSource.toString()}" />
        <link rel="stylesheet" href="${cssModuleSource.toString()}" />
    </head>
    <body>
        <div id="root" />
        <script nonce="${nonce}" src="${scriptSource.toString()}"></script>
    </body>
    </html>`

    return {
        sourcegraphVSCodeSearchSidebarAPI,
    }
}

export function initializeExtensionHostWebview({
    extensionUri,
    webviewView,
    sourcegraphVSCodeExtensionAPI,
}: SourcegraphWebviewConfig & {
    webviewView: vscode.WebviewView
}): {
    sourcegraphVSCodeExtensionHostAPI: Comlink.Remote<SourcegraphVSCodeExtensionHostAPI>
} {
    webviewView.webview.options = {
        enableScripts: true,
    }

    const webviewPath = vscode.Uri.joinPath(extensionUri, 'dist', 'webview')

    const scriptSource = webviewView.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'extensionHost.js'))

    const styleSource = webviewView.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'style.css'))

    const { proxy, expose, panelId } = createEndpointsForWebview(webviewView)

    // Get a proxy for the Sourcegraph Webview API to communicate with the Webview.
    const sourcegraphVSCodeExtensionHostAPI = Comlink.wrap<SourcegraphVSCodeExtensionHostAPI>(proxy)

    // Expose the Sourcegraph VS Code Extension API to the Webview.
    Comlink.expose(sourcegraphVSCodeExtensionAPI, expose)

    // Specific scripts to run using nonce
    const nonce = getNonce()

    // Apply Content-Security-Policy
    // panel.webview.cspSource comes from the webview object
    webviewView.webview.html = `<!DOCTYPE html>
    <html lang="en" data-panel-id="${panelId}" data-instance-url=${endpointSetting()}>
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src vscode-resource: vscode-webview: data: https:; script-src blob: vscode-webview: 'nonce-${nonce}'; style-src data: vscode-resource: ${
        webviewView.webview.cspSource
    } http: https: data:; connect-src 'self' http: https:; font-src vscode-resource: vscode-webview: https:;">
        <title>Sourcegraph Extension Host</title>
        <link rel="stylesheet" href="${styleSource.toString()}" />

    </head>
    <body>
        <div id="root" />
        <script nonce="${nonce}" src="${scriptSource.toString()}"></script>
    </body>
    </html>`

    return { sourcegraphVSCodeExtensionHostAPI }
}

export function getNonce(): string {
    let text = ''
    const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
    for (let index = 0; index < 32; index++) {
        text += possible.charAt(Math.floor(Math.random() * possible.length))
    }
    return text
}
