import * as vscode from 'vscode'

export function activate(context: vscode.ExtensionContext): void {
    console.log('activating Sourcegraph VS Code extension!!!!!', { context })

    // TODO settings management. Should settings management be a concern of general state manager?
    // Perhaps a top level state should be "context-invalidated", which occurs whenever url or accessToken settings change.
    // In this state. search panel must be closed and the search sidebar explains that the user should reload to use the ext
    // (display a reload button; don't auto reload like we currently do)

    /**
     *
     *
     *
     *                                   ┌──────────────────────────┐
     *                                   │                          │
     *                       ┌───────────┤ VS Code extension "core" ├───────────────┐
     *                       │           │          (HERE)          │               │
     *                       │           └──────────────────────────┘               │
     *                       │                                                      │
     *         ┌─────────────▼────────────┐                          ┌──────────────▼───────────┐
     *         │                          │                          │                          │
     *     ┌───┤ "Search sidebar" webview │                          │  "Search panel" webview  │
     *     │   │                          │                          │                          │
     *     │   └──────────────────────────┘                          └──────────────────────────┘
     *     │
     *    ┌▼───────────────────────────┐
     *    │                            │
     *    │ Extension host Web Worker  │
     *    │                            │
     *    └────────────────────────────┘
     *
     *
     * NOTES:
     * - State should live in the core, events flow from either VS Code extension API events (e.g command palette) or webviews (e.g. search box)
     * When the event originates in a webview, it calls the ExtensionCoreAPI dispatch method. Webviews subscribe to the ExtensionCoreAPI
     * `getState()` Observable, which emits on both state and context (e.g. search query) changes.
     * - A reducer may suffice for our needs (no XState?)
     *
     *
     */
}
