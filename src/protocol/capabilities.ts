import { CommandClientCapabilities, ExecuteCommandOptions } from './command'
import { ConfigurationClientCapabilities } from './configuration'
import { ContributionClientCapabilities, ContributionServerCapabilities } from './contribution'
import { DecorationClientCapabilities, DecorationServerCapabilities } from './decoration'
import { ImplementationClientCapabilities, ImplementationServerCapabilities } from './implementation'
import { TextDocumentClientCapabilities, TextDocumentSyncOptions } from './textDocument'
import { TypeDefinitionClientCapabilities, TypeDefinitionServerCapabilities } from './typeDefinition'

/**
 * Defines the capabilities provided by the client.
 */
// tslint:disable-next-line:class-name
export interface _ClientCapabilities {
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
    CommandClientCapabilities &
    ImplementationClientCapabilities &
    TypeDefinitionClientCapabilities &
    ConfigurationClientCapabilities &
    ContributionClientCapabilities &
    DecorationClientCapabilities

/**
 * Defines the capabilities provided by a language
 * server.
 */
// tslint:disable-next-line:class-name
export interface _ServerCapabilities {
    /**
     * Defines how text documents are synced.
     */
    textDocumentSync?: TextDocumentSyncOptions
    /**
     * The server provides hover support.
     */
    hoverProvider?: boolean
    /**
     * The server provides goto definition support.
     */
    definitionProvider?: boolean
    /**
     * The server provides find references support.
     */
    referencesProvider?: boolean
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
    ContributionServerCapabilities &
    DecorationServerCapabilities
