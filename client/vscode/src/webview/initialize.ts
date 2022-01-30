import * as Comlink from 'comlink'
import { Observable } from 'rxjs'
import { filter, first } from 'rxjs/operators'
import * as vscode from 'vscode'

import { ExtensionCoreAPI, SearchPanelAPI, SearchSidebarAPI } from '../contract'

import { createEndpointsForWebview } from './comlink/extensionEndpoint'

interface SourcegraphWebviewConfig {
    extensionUri: vscode.Uri
    extensionCoreAPI: ExtensionCoreAPI
}

export async function initializeSearchPanelWebview({
    extensionUri,
    extensionCoreAPI,
    initializedPanelIDs,
}: SourcegraphWebviewConfig & {
    initializedPanelIDs: Observable<string>
}): Promise<{
    searchPanelAPI: Comlink.Remote<SearchPanelAPI>
    webviewPanel: vscode.WebviewPanel
}> {
    const panel = vscode.window.createWebviewPanel('sourcegraphSearch', 'Sourcegraph', vscode.ViewColumn.One, {
        enableScripts: true,
        retainContextWhenHidden: true,
        enableFindWidget: true,
        localResourceRoots: [vscode.Uri.joinPath(extensionUri, 'dist', 'webview')],
    })

    const webviewPath = vscode.Uri.joinPath(extensionUri, 'dist', 'webview')

    const scriptSource = panel.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'searchPanel.js'))
    const cssModuleSource = panel.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'searchPanel.css'))
    const styleSource = panel.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'style.css'))
    const codiconFontSource = panel.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'codicon.ttf'))

    const { proxy, expose, panelId } = createEndpointsForWebview(panel)

    // Wait for the webview to initialize or else messages will be dropped
    const hasInitialized = initializedPanelIDs
        .pipe(
            filter(initializedPanelId => initializedPanelId === panelId),
            first()
        )
        .toPromise()

    // Get a proxy for the search panel API to communicate with the Webview.
    const searchPanelAPI = Comlink.wrap<SearchPanelAPI>(proxy)

    // Expose the "Core" extension API to the Webview.
    Comlink.expose(extensionCoreAPI, expose)

    // Use a nonce to only allow specific scripts to be run
    const nonce = getNonce()

    panel.iconPath = vscode.Uri.joinPath(extensionUri, 'images', 'logo.svg')

    // Apply Content-Security-Policy
    // panel.webview.cspSource comes from the webview object
    // debt: load codicon ourselves.
    panel.webview.html = `<!DOCTYPE html>
    <html lang="en" data-panel-id="${panelId}">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <style nonce="${nonce}">
            @font-face {
                font-family: 'codicon';
                src: url(${codiconFontSource.toString()})
            }
        </style>
        <meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src data: vscode-resource: vscode-webview: https:; script-src 'nonce-${nonce}' vscode-webview:; style-src data: ${
        panel.webview.cspSource
    } vscode-resource: vscode-webview: 'unsafe-inline' http: https: data:; connect-src 'self' vscode-webview: http: https:; frame-src https:; font-src ${
        panel.webview.cspSource
    };">
        <title>Sourcegraph Search</title>
        <link rel="stylesheet" href="${styleSource.toString()}" />
        <link rel="stylesheet" href="${cssModuleSource.toString()}" />
    </head>
    <body class="search-panel">
        <div id="root" />
        <script nonce="${nonce}" src="${scriptSource.toString()}"></script>
    </body>
    </html>`

    await hasInitialized

    return {
        searchPanelAPI,
        webviewPanel: panel,
    }
}

// TODO expand CSP for Sourcegraph extension loading
export function initializeSearchSidebarWebview({
    extensionUri,
    extensionCoreAPI,
    webviewView,
}: SourcegraphWebviewConfig & {
    webviewView: vscode.WebviewView
}): {
    searchSidebarAPI: Comlink.Remote<SearchSidebarAPI>
} {
    webviewView.webview.options = {
        enableScripts: true,
    }

    const webviewPath = vscode.Uri.joinPath(extensionUri, 'dist', 'webview')

    const scriptSource = webviewView.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'searchSidebar.js'))
    const cssModuleSource = webviewView.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'searchSidebar.css'))
    const styleSource = webviewView.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'style.css'))
    const codiconFontSource = webviewView.webview.asWebviewUri(vscode.Uri.joinPath(webviewPath, 'codicon.ttf'))

    const { proxy, expose, panelId } = createEndpointsForWebview(webviewView)

    // Get a proxy for the Sourcegraph Webview API to communicate with the Webview.
    const searchSidebarAPI = Comlink.wrap<SearchSidebarAPI>(proxy)

    // Expose the Sourcegraph VS Code Extension API to the Webview.
    Comlink.expose(extensionCoreAPI, expose)

    // Specific scripts to run using nonce
    const nonce = getNonce()

    // Apply Content-Security-Policy
    // panel.webview.cspSource comes from the webview object
    // debt: load codicon ourselves.
    webviewView.webview.html = `<!DOCTYPE html>
    <html lang="en" data-panel-id="${panelId}">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <style nonce="${nonce}">
            @font-face {
                font-family: 'codicon';
                src: url(${codiconFontSource.toString()})
            }
        </style>
        <meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src data: vscode-webview: vscode-resource: https:; script-src 'nonce-${nonce}' vscode-webview:; style-src data: ${
        webviewView.webview.cspSource
    } vscode-resource: http: https: data:; connect-src 'self' http: https:; font-src ${webviewView.webview.cspSource};">
        <title>Sourcegraph Search</title>
        <link rel="stylesheet" href="${styleSource.toString()}" />
        <link rel="stylesheet" href="${cssModuleSource.toString()}" />
    </head>
    <body class="search-sidebar">
        <div id="root" />
        <script nonce="${nonce}" src="${scriptSource.toString()}"></script>
    </body>
    </html>`

    return {
        searchSidebarAPI,
    }
}

export function getNonce(): string {
    let text = ''
    const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
    for (let index = 0; index < 32; index++) {
        text += possible.charAt(Math.floor(Math.random() * possible.length))
    }
    return text
}
