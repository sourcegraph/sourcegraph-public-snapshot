import { releaseProxy } from 'comlink'
import { Observable } from 'rxjs'
import * as vscode from 'vscode'

import { ExtensionCoreAPI } from '../contract'

import { initializeSearchPanelWebview, initializeSearchSidebarWebview } from './initialize'

export function registerWebviews({
    context,
    extensionCoreAPI,
    initializedPanelIDs,
}: {
    context: vscode.ExtensionContext
    extensionCoreAPI: ExtensionCoreAPI
    initializedPanelIDs: Observable<string>
}): void {
    // Track current active webview panel to make sure only one panel exists at a time
    let currentActiveWebviewPanel: vscode.WebviewPanel | undefined
    let searchSidebarWebviewView: vscode.WebviewView | undefined

    // TODO if remote files are open from previous session, we need
    // to focus search sidebar to activate code intel (load extension host),
    // and to do that we need to make sourcegraph:// file opening an activation event.

    // Open Sourcegraph search tab on `sourcegraph.search` command.
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.search', async () => {
            // Focus search sidebar in case this command was the activation event,
            // as opposed to visibiilty of sidebar.
            if (!searchSidebarWebviewView) {
                focusSearchSidebar()
            }

            if (currentActiveWebviewPanel) {
                currentActiveWebviewPanel.reveal()
            } else {
                // sourcegraphSettings.refreshSettings()

                const { searchPanelAPI, webviewPanel } = await initializeSearchPanelWebview({
                    extensionUri: context.extensionUri,
                    extensionCoreAPI,
                    initializedPanelIDs,
                })

                currentActiveWebviewPanel = webviewPanel

                webviewPanel.onDidChangeViewState(() => {
                    if (webviewPanel.visible) {
                        focusSearchSidebar()
                    } else {
                        // TODO emit event (should go to idle state if not remote browsing)
                    }
                })

                webviewPanel.onDidDispose(() => {
                    searchPanelAPI[releaseProxy]()
                    currentActiveWebviewPanel = undefined
                    // Ideally focus last used sidebar tab on search panel close. In lieu of that,
                    // just focus the file explorer if the search sidebar is currently focused.
                    if (searchSidebarWebviewView?.visible) {
                        focusFileExplorer()
                    }
                })
            }
        })
    )

    context.subscriptions.push(
        vscode.window.registerWebviewViewProvider(
            'sourcegraph.searchSidebar',
            {
                // This typically will be called only once since `retainContextWhenHidden` is set to `true`.
                resolveWebviewView: (webviewView, _context, _token) => {
                    const { searchSidebarAPI } = initializeSearchSidebarWebview({
                        extensionUri: context.extensionUri,
                        extensionCoreAPI,
                        webviewView,
                    })
                    searchSidebarWebviewView = webviewView
                    // Initialize search panel.
                    openSearchPanelCommand()

                    // Bring search panel back if it was previously closed on sidebar visibility change
                    webviewView.onDidChangeVisibility(() => {
                        if (webviewView.visible) {
                            openSearchPanelCommand()
                        }
                    })
                    webviewView.onDidDispose(() => {
                        searchSidebarAPI[releaseProxy]()
                    })
                },
            },
            { webviewOptions: { retainContextWhenHidden: true } }
        )
    )
}

function openSearchPanelCommand(): void {
    vscode.commands.executeCommand('sourcegraph.search').then(
        () => {},
        error => {
            console.error(error)
        }
    )
}

function focusSearchSidebar(): void {
    vscode.commands.executeCommand('sourcegraph.searchSidebar.focus').then(
        () => {},
        error => {
            console.error(error)
        }
    )
}

function focusFileExplorer(): void {
    vscode.commands.executeCommand('workbench.view.explorer').then(
        () => {},
        error => {
            console.error(error)
        }
    )
}
