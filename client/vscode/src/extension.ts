import 'cross-fetch/polyfill'
import { releaseProxy } from 'comlink'
import { of, ReplaySubject } from 'rxjs'
import vscode from 'vscode'

import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'

import { invalidateClient, requestGraphQLFromVSCode } from './backend/requestGraphQl'
import { initializeSourcegraphSettings } from './backend/settings'
import { toSourcegraphLanguage } from './code-intel/languages'
import { SourcegraphDefinitionProvider } from './code-intel/SourcegraphDefinitionProvider'
import { SourcegraphHoverProvider } from './code-intel/SourcegraphHoverProvider'
import { SourcegraphReferenceProvider } from './code-intel/SourcegraphReferenceProvider'
import { openSourcegraphUriCommand } from './commands.ts/openSourcegraphUriCommand'
import { FilesTreeDataProvider } from './file-system/FilesTreeDataProvider'
import { SourcegraphFileSystemProvider } from './file-system/SourcegraphFileSystemProvider'
import { SourcegraphUri } from './file-system/SourcegraphUri'
import { log } from './log'
import { endpointHostnameSetting, endpointSetting } from './settings/endpointSetting'
import { SourcegraphVSCodeExtensionAPI } from './webview/contract'
import {
    initializeExtensionHostWebview,
    initializeSearchPanelWebview,
    initializeSearchSidebarWebview,
} from './webview/initialize'
import { createSearchSidebarMediator } from './webview/search-sidebar/mediator'

export function activate(context: vscode.ExtensionContext): void {
    // TODO: Close all editors (search panel and remote files) and restart Sourcegraph extension host
    // any time sourcegraph url or TODO access token change to reduce risk of data leaks in logging.
    // Pass this to GraphQL client to avoid making requests to the new instance before restarting VS Code.
    const initialSourcegraphUrl = endpointSetting()
    const instanceHostname = endpointHostnameSetting()

    vscode.workspace.onDidChangeConfiguration(event => {
        if (event.affectsConfiguration('sourcegraph.url')) {
            const newSourcegraphUrl = vscode.workspace.getConfiguration('sourcegraph').get('url')
            if (initialSourcegraphUrl !== newSourcegraphUrl) {
                invalidateClient()

                for (const subscription of context.subscriptions) {
                    subscription.dispose()
                }

                vscode.window
                    .showInformationMessage('Restart VS Code to use the Sourcegraph extension after URL change.')
                    .then(
                        () => {},
                        () => {}
                    )
                // TODO close editors from different instance.
                // fs.purge()
                // TODO Also validate that the extension host only adds documents from the current instance (explicit check, less likely to
                // be an issue but doesn't hurt to be safe).
                // Close all search tabs!
            }
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
        vscode.window.onDidChangeActiveTextEditor(editor => files.didFocus(editor?.document.uri))
    )
    files.didFocus(vscode.window.activeTextEditor?.document.uri).then(
        () => {},
        () => {}
    )

    context.subscriptions.push(
        vscode.commands.registerCommand('extension.openFile', async uri => {
            if (typeof uri === 'string') {
                await openSourcegraphUriCommand(fs, SourcegraphUri.parse(uri))
            } else {
                // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
                log.error(`extension.openFile(${uri}) argument is not a string`)
            }
        })
    )

    // TODO copy from existing extension.
    // context.subscriptions.push(registerSourcegraphGitCommands())

    const sourcegraphSettings = initializeSourcegraphSettings(context.subscriptions)

    // Create sidebar mediator to facilitate communication between search webviews and sidebar
    const searchSidebarMediator = createSearchSidebarMediator(context.subscriptions)

    // Replay subject with large buffer size just in case panels are opened in quick succession.
    const initializedPanelIDs = new ReplaySubject<string>(7)

    const sourcegraphVSCodeExtensionAPI: SourcegraphVSCodeExtensionAPI = {
        requestGraphQL: requestGraphQLFromVSCode,
        getSettings: () => proxySubscribable(sourcegraphSettings.settings),
        ping: () => proxySubscribable(of('pong')),

        observeActiveWebviewQueryState: searchSidebarMediator.observeActiveWebviewQueryState,
        observeActiveWebviewDynamicFilters: searchSidebarMediator.observeActiveWebviewDynamicFilters,
        setActiveWebviewQueryState: searchSidebarMediator.setActiveWebviewQueryState,
        submitActiveWebviewSearch: searchSidebarMediator.submitActiveWebviewSearch,

        getInstanceHostname: () => instanceHostname,
        panelInitialized: panelId => initializedPanelIDs.next(panelId),
        openFile: (uri: string) => openSourcegraphUriCommand(fs, SourcegraphUri.parse(uri)),

        openSearchPanel: () => vscode.commands.executeCommand('sourcegraph.search'),
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
                resolveWebviewView: (webviewView, _context, _token) => {
                    const { sourcegraphVSCodeSearchSidebarAPI } = initializeSearchSidebarWebview({
                        extensionPath: context.extensionPath,
                        sourcegraphVSCodeExtensionAPI,
                        webviewView,
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
                        extensionPath: context.extensionPath,
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
}
