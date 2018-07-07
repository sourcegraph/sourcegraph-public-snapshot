import { SymbolKind } from 'vscode-languageserver-types'

/**
 * Workspace specific client capabilities.
 */
export interface WorkspaceClientCapabilities {
    /**
     * The client supports applying batch edits
     * to the workspace by supporting the request
     * 'workspace/applyEdit'
     */
    applyEdit?: boolean

    /**
     * Capabilities specific to `WorkspaceEdit`s
     */
    workspaceEdit?: {
        /**
         * The client supports versioned document changes in `WorkspaceEdit`s
         */
        documentChanges?: boolean
    }

    /**
     * Capabilities specific to the `workspace/didChangeConfiguration` notification.
     */
    didChangeConfiguration?: {
        /**
         * Did change configuration notification supports dynamic registration.
         */
        dynamicRegistration?: boolean
    }

    /**
     * Capabilities specific to the `workspace/didChangeWatchedFiles` notification.
     */
    didChangeWatchedFiles?: {
        /**
         * Did change watched files notification supports dynamic registration.
         */
        dynamicRegistration?: boolean
    }

    /**
     * Capabilities specific to the `workspace/symbol` request.
     */
    symbol?: {
        /**
         * Symbol request supports dynamic registration.
         */
        dynamicRegistration?: boolean

        /**
         * Specific capabilities for the `SymbolKind` in the `workspace/symbol` request.
         */
        symbolKind?: {
            /**
             * The symbol kind values the client supports. When this
             * property exists the client also guarantees that it will
             * handle values outside its set gracefully and falls back
             * to a default value when unknown.
             *
             * If this property is not present the client only supports
             * the symbol kinds from `File` to `Array` as defined in
             * the initial version of the protocol.
             */
            valueSet?: SymbolKind[]
        }
    }

    /**
     * Capabilities specific to the `workspace/executeCommand` request.
     */
    executeCommand?: {
        /**
         * Execute command supports dynamic registration.
         */
        dynamicRegistration?: boolean
    }
}
