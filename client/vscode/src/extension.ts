import 'cross-fetch/polyfill'
import { releaseProxy } from 'comlink'
import { of, ReplaySubject } from 'rxjs'
import vscode from 'vscode'

import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import { GraphQLResult } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQLFromVSCode } from './backend/requestGraphQl'
import { initializeSourcegraphSettings } from './backend/settings'
import { SourcegraphVSCodeExtensionAPI } from './webview/contract'
import { initializeSearchPanelWebview, initializeSearchSidebarWebview } from './webview/initialize'
import { createSearchSidebarMediator } from './webview/search-sidebar/mediator'

export function activate(context: vscode.ExtensionContext): void {
    // TODO: reload whole extension any time sourcegrap+h url or access token change
    // (reduce risk of data leaks in logging)
    // Only allow files from the current SG instance in the extension host. (query param in Sourcegraph URI?)

    const sourcegraphSettings = initializeSourcegraphSettings(requestGraphQLFromVSCode, context.subscriptions)

    // Create sidebar mediator to facilitate communication between search webviews and sidebar
    const searchSidebarMediator = createSearchSidebarMediator(context.subscriptions)

    // Replay subject with large buffer size just in case panels are opened in quick succession.
    const initializedPanelIDs = new ReplaySubject<string>(7)

    const sourcegraphVSCodeExtensionAPI: SourcegraphVSCodeExtensionAPI = {
        requestGraphQL: async (request: string, variables: any): Promise<GraphQLResult<any>> =>
            requestGraphQLFromVSCode(request, variables),
        getSettings: () => proxySubscribable(sourcegraphSettings.settings),
        ping: () => proxySubscribable(of('pong')),

        observeActiveWebviewQueryState: searchSidebarMediator.observeActiveWebviewQueryState,
        observeActiveWebviewDynamicFilters: searchSidebarMediator.observeActiveWebviewDynamicFilters,
        setActiveWebviewQueryState: searchSidebarMediator.setActiveWebviewQueryState,
        submitActiveWebviewSearch: searchSidebarMediator.submitActiveWebviewSearch,

        panelInitialized: panelId => initializedPanelIDs.next(panelId),
    }

    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.search', async () => {
            sourcegraphSettings.refreshSettings()

            const { sourcegraphVSCodeSearchWebviewAPI, webviewPanel } = await initializeSearchPanelWebview({
                extensionPath: context.extensionPath,
                sourcegraphVSCodeExtensionAPI,
                initializedPanelIDs,
            })

            searchSidebarMediator.addSearchWebviewPanel(webviewPanel, sourcegraphVSCodeSearchWebviewAPI)

            webviewPanel.onDidDispose(() => {
                sourcegraphVSCodeSearchWebviewAPI[releaseProxy]()
            })
        })
    )

    // Trigger initialization of extension host, bring search sidebar into view.
    vscode.commands.executeCommand('sourcegraph.searchSidebar.focus').then(
        () => {},
        error => {
            console.error(error)
        }
    )

    context.subscriptions.push(
        vscode.window.registerWebviewViewProvider(
            'sourcegraph.searchSidebar',
            {
                resolveWebviewView: (webview, _context, _token) => {
                    const { sourcegraphVSCodeSearchSidebarAPI } = initializeSearchSidebarWebview({
                        extensionPath: context.extensionPath,
                        sourcegraphVSCodeExtensionAPI,
                        webviewView: webview,
                    })

                    webview.onDidDispose(() => {
                        sourcegraphVSCodeSearchSidebarAPI[releaseProxy]()
                    })
                },
            },
            { webviewOptions: { retainContextWhenHidden: true } }
        )
    )

    // TODO Sourcegraph extensions
    context.subscriptions.push(
        vscode.window.registerWebviewViewProvider(
            'sourcegraph.extensionHost',
            {
                resolveWebviewView: (webviewView, _context, _token) => {
                    webviewView.webview.options = {
                        enableScripts: true,
                    }

                    webviewView.webview.html = `<!DOCTYPE html>
                <html lang="en">
                <head>
                    <meta charset="UTF-8">
                    <meta name="viewport" content="width=device-width, initial-scale=1.0">
                    <title>Sourcegraph Extension host</title>
                </head>
                <body>
                    <div id="root">
                    <h1>Sourcegraph Extension host</h1>
                    <p>Testing</p>
                    </div>
                </body>
                </html>`
                },
            },
            {
                webviewOptions: {
                    retainContextWhenHidden: true,
                },
            }
        )
    )
}
