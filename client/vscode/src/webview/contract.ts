/**
 * Sourcegraph VS Code methods exposed to Webviews
 *
 * Note: this API object lives in the VS Code extension host process
 */
export interface SourcegraphVSCodeExtensionAPI {
    ping: () => 'pong!'
}

/**
 * Webview methods exposed to the Sourcegraph VS Code extension.
 */
export interface SourcegraphVSCodeWebviewAPI {
    setRoute: (route: 'search') => void
}
