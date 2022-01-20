import { releaseProxy } from 'comlink'
import { of, ReplaySubject } from 'rxjs'
import * as vscode from 'vscode'

import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'

import { ExtensionCoreAPI } from './contract'
import { createVSCEStateMachine } from './state'
import { initializeSearchPanelWebview, initializeSearchSidebarWebview } from './webview/initialize'

// Sourcegraph VS Code extension architecture
// -----
//
//                                   ┌──────────────────────────┐
//                                   │  env: Node OR Web Worker │
//                       ┌───────────┤ VS Code extension "Core" ├───────────────┐
//                       │           │          (HERE)          │               │
//                       │           └──────────────────────────┘               │
//                       │                                                      │
//         ┌─────────────▼────────────┐                          ┌──────────────▼───────────┐
//         │         env: Web         │                          │          env: Web        │
//     ┌───┤ "search sidebar" webview │                          │  "search panel" webview  │
//     │   │                          │                          │                          │
//     │   └──────────────────────────┘                          └──────────────────────────┘
//     │
//    ┌▼───────────────────────────┐
//    │       env: Web Worker      │
//    │ Sourcegraph Extension host │
//    │                            │
//    └────────────────────────────┘
//
// - See './state.ts' for documentation on state management.
//   - One state machine that lives in Core
// - See './contract.ts' to see the APIs for the three main components:
//   - Core, search sidebar, and search panel.
//   - The extension host API is exposed through the search sidebar.
// - See (TODO) for documentation on _how_ communication between contexts works.
//    It is _not_ important to understand this layer to add features to the
//    VS Code extension (that's why it exists, after all).

export function activate(context: vscode.ExtensionContext): void {
    // TODO initialize VS Code settings
    // TODO initialize Sourcegraph settings

    // Initialize core state machine.
    const stateMachine = createVSCEStateMachine()
    const subscription = stateMachine.observeState().subscribe(state => {
        console.log({ state })
    })
    context.subscriptions.push({
        dispose: () => {
            subscription.unsubscribe()
        },
    })
    stateMachine.emit({ type: 'submit_search_query' })

    // Replay subject with large buffer size just in case panels are opened in quick succession.
    const initializedPanelIDs = new ReplaySubject<string>(7)

    const extensionCoreAPI: ExtensionCoreAPI = {
        ping: () => proxySubscribable(of('pong')),
        panelInitialized: panelId => initializedPanelIDs.next(panelId),
    }

    // Track current active webview panel to make sure only one panel exists at a time
    let currentActiveWebviewPanel: vscode.WebviewPanel | undefined
    let searchSidebarWebviewView: vscode.WebviewView | undefined

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
