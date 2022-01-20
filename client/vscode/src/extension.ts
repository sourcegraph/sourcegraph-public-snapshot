import { of, ReplaySubject } from 'rxjs'
import * as vscode from 'vscode'

import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'

import { ExtensionCoreAPI } from './contract'
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
// - See (TODO) for documentation on _how_ communication between contexts works.
//    It is _not_ important to understand this layer to add features to the
//    VS Code extension (that's why it exists, after all).

export function activate(context: vscode.ExtensionContext): void {
    // TODO initialize VS Code settings
    // TODO initialize Sourcegraph settings

    // Initialize core state machine.
    const stateMachine = createVSCEStateMachine()
    stateMachine.emit({ type: 'submit_search_query' })

    // Replay subject with large buffer size just in case panels are opened in quick succession.
    // For search panel webview to signal that it is ready for messages.
    const initializedPanelIDs = new ReplaySubject<string>(7)

    const extensionCoreAPI: ExtensionCoreAPI = {
        ping: () => proxySubscribable(of('pong')),
        panelInitialized: panelId => initializedPanelIDs.next(panelId),
        observeState: () => proxySubscribable(stateMachine.observeState()),
    }

    registerWebviews({ context, extensionCoreAPI, initializedPanelIDs })
}
