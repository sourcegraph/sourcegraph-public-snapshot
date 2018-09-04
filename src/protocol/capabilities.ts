import { CodeLensOptions } from './codeLens'
import { ColorClientCapabilities, ColorServerCapabilities } from './color'
import { ExecuteCommandOptions } from './command'
import { CompletionOptions } from './completion'
import { ConfigurationClientCapabilities } from './configuration'
import { ContributionClientCapabilities, ContributionServerCapabilities } from './contribution'
import { DecorationClientCapabilities, DecorationServerCapabilities } from './decoration'
import { DocumentLinkOptions } from './documentLink'
import { ImplementationClientCapabilities, ImplementationServerCapabilities } from './implementation'
import { SignatureHelpOptions } from './signatureHelp'
import { TextDocumentClientCapabilities, TextDocumentSyncKind, TextDocumentSyncOptions } from './textDocument'
import { TypeDefinitionClientCapabilities, TypeDefinitionServerCapabilities } from './typeDefinition'
import { WorkspaceClientCapabilities } from './workspace'

/**
 * Defines the capabilities provided by the client.
 */
// tslint:disable-next-line:class-name
export interface _ClientCapabilities {
    /**
     * Workspace specific client capabilities.
     */
    workspace?: WorkspaceClientCapabilities

    /**
     * Text document specific client capabilities.
     */
    textDocument?: TextDocumentClientCapabilities

    /**
     * Experimental client capabilities.
     */
    experimental?: any
}

export type ClientCapabilities = _ClientCapabilities &
    ImplementationClientCapabilities &
    TypeDefinitionClientCapabilities &
    ConfigurationClientCapabilities &
    ColorClientCapabilities &
    ContributionClientCapabilities &
    DecorationClientCapabilities

/**
 * Defines the capabilities provided by a language
 * server.
 */
// tslint:disable-next-line:class-name
export interface _ServerCapabilities {
    /**
     * Defines how text documents are synced. Is either a detailed structure defining each notification or
     * for backwards compatibility the TextDocumentSyncKind number.
     */
    textDocumentSync?: TextDocumentSyncOptions | TextDocumentSyncKind
    /**
     * The server provides hover support.
     */
    hoverProvider?: boolean
    /**
     * The server provides completion support.
     */
    completionProvider?: CompletionOptions
    /**
     * The server provides signature help support.
     */
    signatureHelpProvider?: SignatureHelpOptions
    /**
     * The server provides goto definition support.
     */
    definitionProvider?: boolean
    /**
     * The server provides find references support.
     */
    referencesProvider?: boolean
    /**
     * The server provides document highlight support.
     */
    documentHighlightProvider?: boolean
    /**
     * The server provides document symbol support.
     */
    documentSymbolProvider?: boolean
    /**
     * The server provides workspace symbol support.
     */
    workspaceSymbolProvider?: boolean
    /**
     * The server provides code actions.
     */
    codeActionProvider?: boolean
    /**
     * The server provides code lens.
     */
    codeLensProvider?: CodeLensOptions
    /**
     * The server provides document formatting.
     */
    documentFormattingProvider?: boolean
    /**
     * The server provides document range formatting.
     */
    documentRangeFormattingProvider?: boolean
    /**
     * The server provides document formatting on typing.
     */
    documentOnTypeFormattingProvider?: {
        /**
         * A character on which formatting should be triggered, like `}`.
         */
        firstTriggerCharacter: string
        /**
         * More trigger characters.
         */
        moreTriggerCharacter?: string[]
    }
    /**
     * The server provides rename support.
     */
    renameProvider?: boolean
    /**
     * The server provides document link support.
     */
    documentLinkProvider?: DocumentLinkOptions
    /**
     * The server provides execute command support.
     */
    executeCommandProvider?: ExecuteCommandOptions
    /**
     * Experimental server capabilities.
     */
    experimental?: any
}

export type ServerCapabilities = _ServerCapabilities &
    ImplementationServerCapabilities &
    TypeDefinitionServerCapabilities &
    ColorServerCapabilities &
    ContributionServerCapabilities &
    DecorationServerCapabilities
