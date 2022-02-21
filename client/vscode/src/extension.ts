import 'cross-fetch/polyfill'
import { of, ReplaySubject } from 'rxjs'
import * as vscode from 'vscode'

import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'

import { observeAuthenticatedUser } from './backend/authenticatedUser'
import { requestGraphQLFromVSCode } from './backend/requestGraphQl'
import { initializeSourcegraphSettings } from './backend/sourcegraphSettings'
import { ExtensionCoreAPI } from './contract'
import { updateAccessTokenSetting } from './settings/accessTokenSetting'
import { endpointSetting } from './settings/endpointSetting'
import { invalidateContextOnSettingsChange } from './settings/invalidation'
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
    const stateMachine = createVSCEStateMachine()

    invalidateContextOnSettingsChange({ context, stateMachine })
    const sourcegraphSettings = initializeSourcegraphSettings({ context })
    const authenticatedUser = observeAuthenticatedUser({ context })
    const initialInstanceURL = endpointSetting()

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
        openLink: async uri => {
            await vscode.env.openExternal(vscode.Uri.parse(uri))
        },
        setAccessToken: accessToken => updateAccessTokenSetting(accessToken),
        reloadWindow: () => vscode.commands.executeCommand('workbench.action.reloadWindow'),
    }

    registerWebviews({ context, extensionCoreAPI, initializedPanelIDs, sourcegraphSettings })
    // TODO: registerCodeSharingCommands()
    // TODO: registerCodeIntel()
}
