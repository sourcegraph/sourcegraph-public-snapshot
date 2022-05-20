import 'cross-fetch/polyfill'

import { of, ReplaySubject } from 'rxjs'
import vscode, { env } from 'vscode'

import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import polyfillEventSource from '@sourcegraph/shared/src/polyfills/vendor/eventSource'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'

import { observeAuthenticatedUser } from './backend/authenticatedUser'
import { logEvent } from './backend/eventLogger'
import { initializeInstantVersionNumber } from './backend/instanceVersion'
import { requestGraphQLFromVSCode } from './backend/requestGraphQl'
import { initializeSearchContexts } from './backend/searchContexts'
import { initializeSourcegraphSettings } from './backend/sourcegraphSettings'
import { createStreamSearch } from './backend/streamSearch'
import { ExtensionCoreAPI } from './contract'
import { openSourcegraphUriCommand } from './file-system/commands'
import { initializeSourcegraphFileSystem } from './file-system/initialize'
import { SourcegraphUri } from './file-system/SourcegraphUri'
import { Event } from './graphql-operations'
import { initializeCodeSharingCommands } from './link-commands/initialize'
import { accessTokenSetting, updateAccessTokenSetting } from './settings/accessTokenSetting'
import { endpointRequestHeadersSetting, endpointSetting, updateEndpointSetting } from './settings/endpointSetting'
import { invalidateContextOnSettingsChange } from './settings/invalidation'
import { LocalStorageService, SELECTED_SEARCH_CONTEXT_SPEC_KEY } from './settings/LocalStorageService'
import { watchUninstall } from './settings/uninstall'
import { createVSCEStateMachine, VSCEQueryState } from './state'
import { focusSearchPanel, registerWebviews } from './webview/commands'

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
// - See './webview/comlink' for documentation on _how_ communication between contexts works.
//    It is _not_ important to understand this layer to add features to the
//    VS Code extension (that's why it exists, after all).

export function activate(context: vscode.ExtensionContext): void {
    const localStorageService = new LocalStorageService(context.globalState)
    const stateMachine = createVSCEStateMachine({ localStorageService })
    invalidateContextOnSettingsChange({ context, stateMachine })
    initializeSearchContexts({ localStorageService, stateMachine, context })
    const eventSourceType = initializeInstantVersionNumber(localStorageService)
    const sourcegraphSettings = initializeSourcegraphSettings({ context })
    const authenticatedUser = observeAuthenticatedUser({ context })
    const initialInstanceURL = endpointSetting()

    // Sets global `EventSource` for Node, which is required for streaming search.
    // Used for VS Code web as well to be able to add Authorization header.
    const initialAccessToken = accessTokenSetting()
    // Add custom headers to `EventSource` Authorization header when provided
    const customHeaders = endpointRequestHeadersSetting()
    polyfillEventSource(initialAccessToken ? { Authorization: `token ${initialAccessToken}`, ...customHeaders } : {})
    // Update `EventSource` Authorization header on access token / headers change.
    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration(config => {
            if (
                config.affectsConfiguration('sourcegraph.accessToken') ||
                config.affectsConfiguration('sourcegraph.requestHeaders')
            ) {
                const newAccessToken = accessTokenSetting()
                const newCustomHeaders = endpointRequestHeadersSetting()
                polyfillEventSource(
                    newAccessToken ? { Authorization: `token ${newAccessToken}`, ...newCustomHeaders } : {}
                )
            }
        })
    )
    // For search panel webview to signal that it is ready for messages.
    // Replay subject with large buffer size just in case panels are opened in quick succession.
    const initializedPanelIDs = new ReplaySubject<string>(7)

    // Used to observe search box query state from sidebar
    const sidebarQueryStates = new ReplaySubject<VSCEQueryState>(1)

    const { fs } = initializeSourcegraphFileSystem({ context, initialInstanceURL })
    // Use api endpoint for stream search
    const streamSearch = createStreamSearch({ context, stateMachine, sourcegraphURL: `${initialInstanceURL}/.api` })

    const extensionCoreAPI: ExtensionCoreAPI = {
        panelInitialized: panelId => initializedPanelIDs.next(panelId),
        observeState: () => proxySubscribable(stateMachine.observeState()),
        observePanelQueryState: () => proxySubscribable(sidebarQueryStates.asObservable()),
        emit: event => stateMachine.emit(event),
        requestGraphQL: requestGraphQLFromVSCode,
        observeSourcegraphSettings: () => proxySubscribable(sourcegraphSettings.settings),
        // Debt: converting Promises into Observables for ease of use with
        // `useObservable` hook. Add `usePromise`s hook to fix.
        getAuthenticatedUser: () => proxySubscribable(authenticatedUser),
        getInstanceURL: () => proxySubscribable(of(initialInstanceURL)),
        openSourcegraphFile: (uri: string) => openSourcegraphUriCommand(fs, SourcegraphUri.parse(uri)),
        openLink: (uri: string) => vscode.env.openExternal(vscode.Uri.parse(uri)),
        copyLink: (uri: string) =>
            env.clipboard.writeText(uri).then(() => vscode.window.showInformationMessage('Link Copied!')),
        setAccessToken: accessToken => updateAccessTokenSetting(accessToken),
        setEndpointUri: uri => updateEndpointSetting(uri),
        reloadWindow: () => vscode.commands.executeCommand('workbench.action.reloadWindow'),
        focusSearchPanel,
        streamSearch,
        fetchStreamSuggestions: (query, sourcegraphURL) =>
            // Use api endpoint for stream search
            proxySubscribable(fetchStreamSuggestions(query, `${sourcegraphURL}/.api`)),
        setSelectedSearchContextSpec: spec => {
            stateMachine.emit({ type: 'set_selected_search_context_spec', spec })
            return localStorageService.setValue(SELECTED_SEARCH_CONTEXT_SPEC_KEY, spec)
        },
        setSidebarQueryState: sidebarQueryState => sidebarQueryStates.next(sidebarQueryState),
        getLocalStorageItem: key => localStorageService.getValue(key),
        setLocalStorageItem: (key: string, value: string) => localStorageService.setValue(key, value),
        logEvents: (variables: Event) => logEvent(variables),
        getEventSource: eventSourceType,
    }

    // Also initializes code intel.
    registerWebviews({
        context,
        extensionCoreAPI,
        initializedPanelIDs,
        sourcegraphSettings,
        fs,
        instanceURL: initialInstanceURL,
    })
    initializeCodeSharingCommands(context, eventSourceType, localStorageService)
    watchUninstall(eventSourceType, localStorageService)
}
