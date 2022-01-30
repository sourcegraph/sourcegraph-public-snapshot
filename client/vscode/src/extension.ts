import 'cross-fetch/polyfill'
import { of, ReplaySubject } from 'rxjs'
import vscode, { env } from 'vscode'

import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'

import { observeAuthenticatedUser } from './backend/authenticatedUser'
import { requestGraphQLFromVSCode } from './backend/requestGraphQl'
import { initializeSearchContexts } from './backend/searchContexts'
import { initializeSourcegraphSettings } from './backend/sourcegraphSettings'
import { createStreamSearch } from './backend/streamSearch'
import { ExtensionCoreAPI } from './contract'
import polyfillEventSource from './polyfills/eventSource'
import { accessTokenSetting, updateAccessTokenSetting } from './settings/accessTokenSetting'
import { endpointSetting } from './settings/endpointSetting'
import { invalidateContextOnSettingsChange } from './settings/invalidation'
import { LocalStorageService, SELECTED_SEARCH_CONTEXT_SPEC_KEY } from './settings/LocalStorageService'
import { createVSCEStateMachine } from './state'
import { registerWebviews } from './webview/commands'

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
    const localStorageService = new LocalStorageService(context.workspaceState)
    const stateMachine = createVSCEStateMachine({ localStorageService })

    invalidateContextOnSettingsChange({ context, stateMachine })
    initializeSearchContexts({ localStorageService, stateMachine, context })
    const sourcegraphSettings = initializeSourcegraphSettings({ context })
    const authenticatedUser = observeAuthenticatedUser({ context })
    const initialInstanceURL = endpointSetting()

    // Sets global `EventSource` for Node, which is required for streaming search.
    // Used for VS Code web as well to be able to add Authorization header.
    const initialAccessToken = accessTokenSetting()
    polyfillEventSource(initialAccessToken ? { Authorization: `token ${initialAccessToken}` } : {})
    // Update `EventSource` Authorization header on access token change.
    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration(config => {
            if (config.affectsConfiguration('sourcegraph.accessToken')) {
                const newAccessToken = accessTokenSetting()
                polyfillEventSource(newAccessToken ? { Authorization: `token ${newAccessToken}` } : {})
            }
        })
    )

    // Add state to VS Code context to be used in context keys.
    // Used e.g. by file tree view to only be visible in `remote-browsing` state.
    const subscription = stateMachine.observeState().subscribe(state => {
        vscode.commands.executeCommand('setContext', 'sourcegraph.state', state.status).then(
            () => {},
            () => {}
        )
    })
    context.subscriptions.push({
        dispose: () => subscription.unsubscribe(),
    })

    // For search panel webview to signal that it is ready for messages.
    // Replay subject with large buffer size just in case panels are opened in quick succession.
    const initializedPanelIDs = new ReplaySubject<string>(7)

    const streamSearch = createStreamSearch({ context, stateMachine, sourcegraphURL: initialInstanceURL })

    const extensionCoreAPI: ExtensionCoreAPI = {
        panelInitialized: panelId => initializedPanelIDs.next(panelId),
        observeState: () => proxySubscribable(stateMachine.observeState()),
        emit: event => stateMachine.emit(event),
        requestGraphQL: requestGraphQLFromVSCode,
        observeSourcegraphSettings: () => proxySubscribable(sourcegraphSettings.settings),
        // Debt: converting Promises into Observables for ease of use with
        // `useObservable` hook. Add `usePromise`s hook to fix.
        getAuthenticatedUser: () => proxySubscribable(authenticatedUser),
        getInstanceURL: () => proxySubscribable(of(initialInstanceURL)),
        openLink: (uri: string) => vscode.env.openExternal(vscode.Uri.parse(uri)),
        copyLink: (uri: string) =>
            env.clipboard.writeText(uri).then(() => vscode.window.showInformationMessage('Link Copied!')),
        setAccessToken: accessToken => updateAccessTokenSetting(accessToken),
        reloadWindow: () => vscode.commands.executeCommand('workbench.action.reloadWindow'),
        streamSearch,
        fetchStreamSuggestions: (query, sourcegraphURL) =>
            proxySubscribable(fetchStreamSuggestions(query, sourcegraphURL)),
        setSelectedSearchContextSpec: spec => {
            stateMachine.emit({ type: 'set_selected_search_context_spec', spec })
            return localStorageService.setValue(SELECTED_SEARCH_CONTEXT_SPEC_KEY, spec)
        },
    }

    registerWebviews({ context, extensionCoreAPI, initializedPanelIDs, sourcegraphSettings })
    // TODO: registerCodeSharingCommands()
    // TODO: registerCodeIntel()
}
