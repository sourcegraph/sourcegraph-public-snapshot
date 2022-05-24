import { Observable } from 'rxjs'
import * as vscode from 'vscode'

import { LATEST_VERSION } from '@sourcegraph/shared/src/search/stream'

import { initializeSourcegraphSettings } from '../backend/sourcegraphSettings'
import { initializeCodeIntel } from '../code-intel/initialize'
import { ExtensionCoreAPI } from '../contract'
import { SourcegraphFileSystemProvider } from '../file-system/SourcegraphFileSystemProvider'
import { SearchPatternType } from '../graphql-operations'

import {
    initializeHelpSidebarWebview,
    initializeSearchPanelWebview,
    initializeSearchSidebarWebview,
} from './initialize'

// Track current active webview panel to make sure only one panel exists at a time
let currentSearchPanel: vscode.WebviewPanel | 'initializing' | undefined
let searchSidebarWebviewView: vscode.WebviewView | 'initializing' | undefined

export function registerWebviews({
    context,
    extensionCoreAPI,
    initializedPanelIDs,
    sourcegraphSettings,
    fs,
    instanceURL,
}: {
    context: vscode.ExtensionContext
    extensionCoreAPI: ExtensionCoreAPI
    initializedPanelIDs: Observable<string>
    sourcegraphSettings: ReturnType<typeof initializeSourcegraphSettings>
    fs: SourcegraphFileSystemProvider
    instanceURL: string
}): void {
    // TODO if remote files are open from previous session, we need to focus search sidebar to activate code intel (load extension host)
    /**
     * URI Handler to resolve data sending back from Browser
     */
    const handleUri = async (uri: vscode.Uri): Promise<void> => {
        const token = new URLSearchParams(uri.query).get('code')
        // const returnedNonce = new URLSearchParams(uri.query).get('nonce')
        // TODO: Decrypt token
        // TODO: Match returnedNonce to stored nonce
        if (token && token.length > 8) {
            await vscode.workspace
                .getConfiguration('sourcegraph')
                .update('accessToken', token, vscode.ConfigurationTarget.Global)
            await vscode.window.showInformationMessage('Token has been retreived and updated successfully')
        }
    }
    /**
     * Create URI Handler to resolve data sending back from Browser
     */
    context.subscriptions.push(
        vscode.window.registerUriHandler({
            handleUri,
        })
    )
    /**
     * Create command to open Sourcegraph search tab on `sourcegraph.search`
     * Create webview for the search panel if one has not been created
     */
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.search', async (searchQuery?: string) => {
            // If text selected, submit search for it. Capture selection first.
            const activeEditor = vscode.window.activeTextEditor
            const selection = activeEditor?.selection
            // If searchQuery is provided, ignore selection and use searchQuery instead
            const selectedQuery = searchQuery || activeEditor?.document.getText(selection)
            // Focus search sidebar in case this command was the activation event,
            // as opposed to visibiilty of sidebar.
            if (!searchSidebarWebviewView) {
                focusSearchSidebar()
            }
            // If there is a search panel opened, reveal it, or create one if it doesn't exist
            if (currentSearchPanel && currentSearchPanel !== 'initializing') {
                currentSearchPanel.reveal()
            } else if (!currentSearchPanel) {
                sourcegraphSettings.refreshSettings()
                currentSearchPanel = 'initializing'
                const { webviewPanel, searchPanelAPI } = await initializeSearchPanelWebview({
                    extensionUri: context.extensionUri,
                    extensionCoreAPI,
                    initializedPanelIDs,
                })
                currentSearchPanel = webviewPanel
                webviewPanel.onDidChangeViewState(() => {
                    if (webviewPanel.active) {
                        extensionCoreAPI.emit({ type: 'search_panel_focused' })
                        focusSearchSidebar()
                        searchPanelAPI.focusSearchBox().catch(() => {})
                    }
                    if (webviewPanel.visible) {
                        searchPanelAPI.focusSearchBox().catch(() => {})
                    }
                    if (!webviewPanel.visible) {
                        // TODO emit event (should go to idle state if not remote browsing)
                        extensionCoreAPI.emit({ type: 'search_panel_unfocused' })
                    }
                })
                webviewPanel.onDidDispose(() => {
                    currentSearchPanel = undefined
                    // Ideally focus last used sidebar tab on search panel close. In lieu of that (for v1),
                    // just focus the file explorer if the search sidebar is currently focused.
                    if (searchSidebarWebviewView !== 'initializing' && searchSidebarWebviewView?.visible) {
                        focusFileExplorer()
                    }
                    // Clear search result
                    extensionCoreAPI.emit({ type: 'search_panel_disposed' })
                })
            }
            // Submit search query if selected text is detected or a search query provided from search box
            if (selectedQuery) {
                extensionCoreAPI.streamSearch(selectedQuery, {
                    patternType: SearchPatternType.literal,
                    caseSensitive: false,
                    version: LATEST_VERSION,
                    trace: undefined,
                    sourcegraphURL: instanceURL,
                })
            }
        })
    )
    /**
     * Create webview for the search sidebar
     */
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

                    initializeCodeIntel({ context, fs, searchSidebarAPI })

                    // Bring search panel back if it was previously closed on sidebar visibility change
                    webviewView.onDidChangeVisibility(() => {
                        if (webviewView.visible) {
                            openSearchPanelCommand()
                        }
                    })
                },
            },
            { webviewOptions: { retainContextWhenHidden: true } }
        )
    )
    /**
     * Create webview for the help and feedback sidebar
     */
    context.subscriptions.push(
        vscode.window.registerWebviewViewProvider(
            'sourcegraph.helpSidebar',
            {
                // This typically will be called only once since `retainContextWhenHidden` is set to `true`.
                resolveWebviewView: (webviewView, _context, _token) => {
                    initializeHelpSidebarWebview({
                        extensionUri: context.extensionUri,
                        extensionCoreAPI,
                        webviewView,
                    })
                },
            },
            { webviewOptions: { retainContextWhenHidden: true } }
        )
    )
    /**
     * Clone Remote Git Repos Locally using VS Code Git API
     * Ref: https://github.com/microsoft/vscode/issues/48428
     */
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.gitClone', async () => {
            const editor = vscode.window.activeTextEditor
            if (!editor) {
                throw new Error('No active editor')
            }
            const uri = editor.document.uri.path
            const gitUrl = `https:/${uri.split('@')[0]}.git`
            const vsCodeCloneUrl = `vscode://vscode.git/clone?url=${gitUrl}`
            await vscode.env.openExternal(vscode.Uri.parse(vsCodeCloneUrl))
            // vscode://vscode.git/clone?url=${gitUrl}
        })
    )
    /**
     * This is to open the input box in command pa
     */
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.quickSearch', async () => {
            const query = await vscode.window.showInputBox({
                title: 'Sourcegraph Search',
                placeHolder: 'Example: repo:sourcegraph/* lang:TypeScript -file:test createStreamSearch',
                prompt: 'Enter search query...',
                ignoreFocusOut: true,
            })
            return openSearchPanelCommand(query)
        })
    )
}
/**
 * This is to open Search Panel by running the command
 */
function openSearchPanelCommand(searchQuery?: string): void {
    vscode.commands.executeCommand('sourcegraph.search', searchQuery).then(
        () => {},
        error => {
            console.error(error)
        }
    )
}
/**
 * This is to bring focus to the search sidebar
 */
function focusSearchSidebar(): void {
    vscode.commands.executeCommand('sourcegraph.searchSidebar.focus').then(
        () => {},
        error => {
            console.error(error)
        }
    )
}
/**
 * This is to show the search panel if current on another window
 */
export function focusSearchPanel(): void {
    if (currentSearchPanel && currentSearchPanel !== 'initializing') {
        currentSearchPanel.reveal()
    }
}
/**
 * This is to bring focus to the file explorer
 */
function focusFileExplorer(): void {
    vscode.commands.executeCommand('workbench.view.explorer').then(
        () => {},
        error => {
            console.error(error)
        }
    )
}
