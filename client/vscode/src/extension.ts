import 'cross-fetch/polyfill'
import { releaseProxy } from 'comlink'
import { of, ReplaySubject } from 'rxjs'
import vscode, { env } from 'vscode'

import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'

import { LocalStorageService } from '../localStorageService'

import { invalidateClient, requestGraphQLFromVSCode, currentUserSettings } from './backend/requestGraphQl'
import { initializeSourcegraphSettings } from './backend/settings'
import { toSourcegraphLanguage } from './code-intel/languages'
import { SourcegraphDefinitionProvider } from './code-intel/SourcegraphDefinitionProvider'
import { SourcegraphHoverProvider } from './code-intel/SourcegraphHoverProvider'
import { SourcegraphReferenceProvider } from './code-intel/SourcegraphReferenceProvider'
import { inBrowserActions, openLinkInBrowser } from './commands/node/inBrowserActions'
import { openSourcegraphUriCommand } from './commands/openSourcegraphUriCommand'
import { searchSelection } from './commands/searchSelection'
import { FilesTreeDataProvider } from './file-system/FilesTreeDataProvider'
import { SourcegraphFileSystemProvider } from './file-system/SourcegraphFileSystemProvider'
import { SourcegraphUri } from './file-system/SourcegraphUri'
import { log } from './log'
import { updateAccessTokenSetting } from './settings/accessTokenSetting'
import { updateCorsSetting } from './settings/endpointSetting'
import { LocalRecentSeachProps, SourcegraphVSCodeExtensionAPI } from './webview/contract'
import {
    initializeExtensionHostWebview,
    initializeSearchPanelWebview,
    initializeSearchSidebarWebview,
} from './webview/initialize'
import { createSearchSidebarMediator } from './webview/search-sidebar/mediator'

export function activate(context: vscode.ExtensionContext): void {
    // Initialize the global application manager
    const storageManager = new LocalStorageService(context.workspaceState)
    // TODO: Close all editors (search panel and remote files) and restart Sourcegraph extension host
    // any time sourcegraph url or TODO access token change to reduce risk of data leaks in logging.
    // Pass this to GraphQL client to avoid making requests to the new instance before restarting VS Code.
    const userSettings = currentUserSettings()
    const initialSourcegraphUrl = userSettings.endpoint
    const allLocalSearchHistory = storageManager.getUserLocalSearchHistory()

    vscode.workspace.onDidChangeConfiguration(async event => {
        if (event.affectsConfiguration('sourcegraph.url')) {
            const newSourcegraphUrl = vscode.workspace.getConfiguration('sourcegraph').get('url')
            if (initialSourcegraphUrl !== newSourcegraphUrl) {
                invalidateClient()

                for (const subscription of context.subscriptions) {
                    subscription.dispose()
                }
                // TODO close editors from different instance.
                // fs.purge()
                // TODO Also validate that the extension host only adds documents from the current instance (explicit check, less likely to
                // be an issue but doesn't hurt to be safe).
                // Close all search tabs!
            }
        }
        // Reload VS Code with new settings
        if (event.affectsConfiguration('sourcegraph.url') || event.affectsConfiguration('sourcegraph.accessToken')) {
            await vscode.commands.executeCommand('workbench.action.reloadWindow')
        }
    })

    // Register file-system related functionality.
    const fs = new SourcegraphFileSystemProvider(initialSourcegraphUrl)
    const files = new FilesTreeDataProvider(fs)

    vscode.workspace.registerFileSystemProvider('sourcegraph', fs, { isReadonly: true })

    const filesTreeView = vscode.window.createTreeView<string>('sourcegraph.files', {
        treeDataProvider: files,
        showCollapseAll: true,
    })

    files.setTreeView(filesTreeView)
    context.subscriptions.push(filesTreeView)
    context.subscriptions.push(
        vscode.window.onDidChangeActiveTextEditor(async editor => {
            const vsceUri = editor?.document.uri
            await files.didFocus(vsceUri)
            // if it's a sourcegraph remote file
            // we will add it to local storage as recent file search
            // for easy access in sidebar later
            if (vsceUri && files.isSourcegrapeRemoteFile(vsceUri)) {
                const currentFileHistory = storageManager.getFileHistory()
                const sgUri = fs.sourcegraphUri(vsceUri).uri
                const fileHistorySet = new Set(currentFileHistory).add(sgUri)
                await storageManager.setFileHistory([...fileHistorySet].slice(-9))
            }
        })
    )
    files.didFocus(vscode.window.activeTextEditor?.document.uri).then(
        async () => {},
        () => {}
    )

    const sourcegraphSettings = initializeSourcegraphSettings(context.subscriptions)

    // Create sidebar mediator to facilitate communication between search webviews and sidebar
    const searchSidebarMediator = createSearchSidebarMediator(context.subscriptions)

    // Replay subject with large buffer size just in case panels are opened in quick succession.
    const initializedPanelIDs = new ReplaySubject<string>(7)

    // Track current active webview panel to make sure only one panel exists at a time
    let currentActiveWebviewPanel: vscode.WebviewPanel | undefined

    // Open Sourcegraph search tab
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.search', async () => {
            if (currentActiveWebviewPanel) {
                currentActiveWebviewPanel.reveal()
            } else {
                sourcegraphSettings.refreshSettings()

                const { sourcegraphVSCodeSearchWebviewAPI, webviewPanel } = await initializeSearchPanelWebview({
                    extensionUri: context.extensionUri,
                    sourcegraphVSCodeExtensionAPI,
                    initializedPanelIDs,
                })
                currentActiveWebviewPanel = webviewPanel

                webviewPanel.onDidChangeViewState(async () => {
                    if (webviewPanel.visible) {
                        await vscode.commands.executeCommand('setContext', 'sourcegraph.showFileTree', false)
                    }
                })

                searchSidebarMediator.addSearchWebviewPanel(webviewPanel, sourcegraphVSCodeSearchWebviewAPI)
                await vscode.commands.executeCommand('setContext', 'sourcegraph.activeSearchPanel', false)
                webviewPanel.onDidDispose(async () => {
                    sourcegraphVSCodeSearchWebviewAPI[releaseProxy]()
                    currentActiveWebviewPanel = undefined
                    // Set showFileTree and activeSearchPanel to false
                    await vscode.commands.executeCommand('setContext', 'sourcegraph.activeSearchPanel', false)
                    await vscode.commands.executeCommand('setContext', 'sourcegraph.showFileTree', false)
                    // Close sidebar when user closes the search panel by displaying their explorer instead
                    await vscode.commands.executeCommand('workbench.view.explorer')
                })
            }
        })
    )

    vscode.commands.executeCommand('sourcegraph.searchSidebar.focus').then(
        () => {},
        () => {}
    )

    context.subscriptions.push(
        vscode.window.registerWebviewViewProvider(
            'sourcegraph.searchSidebar',
            {
                resolveWebviewView: (webviewView, _context, _token) => {
                    const { sourcegraphVSCodeSearchSidebarAPI } = initializeSearchSidebarWebview({
                        extensionUri: context.extensionUri,
                        sourcegraphVSCodeExtensionAPI,
                        webviewView,
                    })
                    // Bring search panel back if it was previously closed on sidebar visibility change
                    webviewView.onDidChangeVisibility(async () => {
                        if (webviewView.visible) {
                            await vscode.commands.executeCommand('sourcegraph.search')
                        }
                    })
                    webviewView.onDidDispose(() => {
                        sourcegraphVSCodeSearchSidebarAPI[releaseProxy]()
                    })
                },
            },
            { webviewOptions: { retainContextWhenHidden: true } }
        )
    )

    context.subscriptions.push(
        vscode.window.registerWebviewViewProvider(
            'sourcegraph.extensionHost',
            {
                resolveWebviewView: (webviewView, _context, _token) => {
                    const { sourcegraphVSCodeExtensionHostAPI } = initializeExtensionHostWebview({
                        extensionUri: context.extensionUri,
                        sourcegraphVSCodeExtensionAPI,
                        webviewView,
                    })

                    // TODO: send message to Sourcegraph extension host when instance URL changes to shut it down.

                    // Register language-related features (they depend on Sourcegraph extensions).
                    context.subscriptions.push(
                        vscode.languages.registerDefinitionProvider(
                            { scheme: 'sourcegraph' },
                            new SourcegraphDefinitionProvider(fs, sourcegraphVSCodeExtensionHostAPI)
                        )
                    )
                    context.subscriptions.push(
                        vscode.languages.registerReferenceProvider(
                            { scheme: 'sourcegraph' },
                            new SourcegraphReferenceProvider(fs, sourcegraphVSCodeExtensionHostAPI)
                        )
                    )
                    context.subscriptions.push(
                        vscode.languages.registerHoverProvider(
                            { scheme: 'sourcegraph' },
                            new SourcegraphHoverProvider(fs, sourcegraphVSCodeExtensionHostAPI)
                        )
                    )

                    // TODO remove closed editors/documents

                    vscode.window.onDidChangeActiveTextEditor(editor => {
                        // TODO store previously active editor -> SG viewer so we can remove on change
                        if (editor?.document.uri.scheme === 'sourcegraph') {
                            const text = editor.document.getText()
                            const sourcegraphUri = fs.sourcegraphUri(editor.document.uri)
                            const languageId = toSourcegraphLanguage(editor.document.languageId)

                            const extensionHostUri = makeRepoURI({
                                repoName: sourcegraphUri.repositoryName,
                                revision: sourcegraphUri.revision,
                                filePath: sourcegraphUri.path,
                            })

                            // We'll use the viewerId return value to remove viewer, get/set text decorations.
                            sourcegraphVSCodeExtensionHostAPI
                                .addTextDocumentIfNotExists({
                                    text,
                                    uri: extensionHostUri,
                                    languageId,
                                })
                                .then(() =>
                                    sourcegraphVSCodeExtensionHostAPI.addViewerIfNotExists({
                                        type: 'CodeEditor',
                                        resource: extensionHostUri,
                                        selections: [],
                                        isActive: true,
                                    })
                                )
                                .catch(error => console.error(error))
                        }
                    })
                },
            },
            {
                webviewOptions: {
                    retainContextWhenHidden: true,
                },
            }
        )
    )

    const sourcegraphVSCodeExtensionAPI: SourcegraphVSCodeExtensionAPI = {
        requestGraphQL: requestGraphQLFromVSCode,
        getSettings: () => proxySubscribable(sourcegraphSettings.settings),
        ping: () => proxySubscribable(of('pong')),
        observeActiveWebviewQueryState: searchSidebarMediator.observeActiveWebviewQueryState,
        observeActiveWebviewDynamicFilters: searchSidebarMediator.observeActiveWebviewDynamicFilters,
        setActiveWebviewQueryState: searchSidebarMediator.setActiveWebviewQueryState,
        submitActiveWebviewSearch: searchSidebarMediator.submitActiveWebviewSearch,
        getUserSettings: () => userSettings,
        getLocalSearchHistory: () => allLocalSearchHistory,
        hasAccessToken: () => !!userSettings.token,
        hasValidAccessToken: () => userSettings.validated,
        updateAccessToken: (token: string) => updateAccessTokenSetting(token),
        getInstanceHostname: () => userSettings.host,
        panelInitialized: panelId => initializedPanelIDs.next(panelId),
        // Call from webview's search results
        openFile: (uri: string) => openSourcegraphUriCommand(fs, SourcegraphUri.parse(uri)),
        // Open Links in Browser
        openLink: (uri: string) => openLinkInBrowser(uri),
        copyLink: (uri: string) =>
            env.clipboard.writeText(uri).then(() => vscode.window.showInformationMessage('Link Copied!')),
        openSearchPanel: () => vscode.commands.executeCommand('sourcegraph.search'),
        // Check if on VS Code Desktop or VS Code Web
        onDesktop: () => vscode.env.appHost === 'desktop',
        // Get Cors from Setting
        getCorsSetting: () => userSettings.corsUrl,
        updateCorsUri: (uri: string) => updateCorsSetting(uri),
        // Get last selected search context from Setting
        getLastSelectedSearchContext: () => storageManager.getValue('sg-last-selected-context'),
        updateLastSelectedSearchContext: (spec: string) => storageManager.setValue('sg-last-selected-context', spec),
        // Get last selected search context from Setting
        getLocalRecentSearch: () => allLocalSearchHistory.searches,
        setLocalRecentSearch: (searches: LocalRecentSeachProps[]) => storageManager.setLocalRecentSearch(searches),
        // Get last selected search context from Setting
        getLocalStorageItem: (key: string) => storageManager.getValue(key),
        setLocalStorageItem: (key: string, value: string) => storageManager.setValue(key, value),
        // Show File Tree
        displayFileTree: (setting: boolean) =>
            vscode.commands.executeCommand('setContext', 'sourcegraph.showFileTree', setting),
        hasActivePanel: () => vscode.commands.executeCommand('setContext', 'sourcegraph.activeSearchPanel', true),
    }

    // Commands
    // Open remote Sourcegraph file from remote file tree
    context.subscriptions.push(
        vscode.commands.registerCommand('extension.openFile', async uri => {
            if (typeof uri === 'string') {
                await openSourcegraphUriCommand(fs, SourcegraphUri.parse(uri))
                await vscode.commands.executeCommand('setContext', 'sourcegraph.showFileTree', true)
            } else {
                // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
                log.error(`extension.openRemoteFile(${uri}) argument is not a string`)
            }
        })
    )

    // Open local file or remote Sourcegraph file in browser
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.openInBrowser', async () => {
            await inBrowserActions('open')
        })
    )

    // Copy Sourcegraph link to file
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.copyFileLink', async () => {
            await inBrowserActions('copy')
        })
    )

    // Search Selected on Sourcegraph Web
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.selectionSearchWeb', async () => {
            await searchSelection()
        })
    )

    // Search Selected Text in Sourcegraph Search Tab
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.selectionSearch', async () => {
            const editor = vscode.window.activeTextEditor
            if (!editor) {
                throw new Error('No active editor')
            }
            const selectedQuery = editor.document.getText(editor.selection)
            if (selectedQuery && currentActiveWebviewPanel) {
                currentActiveWebviewPanel.reveal()

                await searchSidebarMediator.submitActiveWebviewSearch({ query: selectedQuery })
            }

            if (selectedQuery && !currentActiveWebviewPanel) {
                sourcegraphSettings.refreshSettings()

                const { sourcegraphVSCodeSearchWebviewAPI, webviewPanel } = await initializeSearchPanelWebview({
                    extensionUri: context.extensionUri,
                    sourcegraphVSCodeExtensionAPI,
                    initializedPanelIDs,
                })

                currentActiveWebviewPanel = webviewPanel

                searchSidebarMediator.addSearchWebviewPanel(webviewPanel, sourcegraphVSCodeSearchWebviewAPI)

                webviewPanel.onDidDispose(() => {
                    sourcegraphVSCodeSearchWebviewAPI[releaseProxy]()
                    currentActiveWebviewPanel = undefined
                })
                await searchSidebarMediator.submitActiveWebviewSearch({ query: selectedQuery })
            }
        })
    )
}
