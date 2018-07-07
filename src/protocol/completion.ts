import { CompletionItem, CompletionList } from 'vscode-languageserver-types'
import { RequestType } from '../jsonrpc2/messages'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from './textDocument'

/**
 * Completion options.
 */
export interface CompletionOptions {
    /**
     * Most tools trigger completion request automatically without explicitly requesting
     * it using a keyboard shortcut (e.g. Ctrl+Space). Typically they do so when the user
     * starts to type an identifier. For example if the user types `c` in a JavaScript file
     * code complete will automatically pop up present `console` besides others as a
     * completion item. Characters that make up identifiers don't need to be listed here.
     *
     * If code complete should automatically be trigger on characters not being valid inside
     * an identifier (for example `.` in JavaScript) list them in `triggerCharacters`.
     */
    triggerCharacters?: string[]

    /**
     * The server provides support to resolve additional
     * information for a completion item.
     */
    resolveProvider?: boolean
}

/**
 * Completion registration options.
 */
export interface CompletionRegistrationOptions extends TextDocumentRegistrationOptions, CompletionOptions {}

/**
 * How a completion was triggered
 */
export namespace CompletionTriggerKind {
    /**
     * Completion was triggered by typing an identifier (24x7 code
     * complete), manual invocation (e.g Ctrl+Space) or via API.
     */
    export const Invoked: 1 = 1

    /**
     * Completion was triggered by a trigger character specified by
     * the `triggerCharacters` properties of the `CompletionRegistrationOptions`.
     */
    export const TriggerCharacter: 2 = 2

    /**
     * Completion was re-triggered as current completion list is incomplete
     */
    export const TriggerForIncompleteCompletions: 3 = 3
}

export type CompletionTriggerKind = 1 | 2 | 3

/**
 * Contains additional information about the context in which a completion request is triggered.
 */
export interface CompletionContext {
    /**
     * How the completion was triggered.
     */
    triggerKind: CompletionTriggerKind

    /**
     * The trigger character (a single character) that has trigger code complete.
     * Is undefined if `triggerKind !== CompletionTriggerKind.TriggerCharacter`
     */
    triggerCharacter?: string
}

/**
 * Completion parameters
 */
export interface CompletionParams extends TextDocumentPositionParams {
    /**
     * The completion context. This is only available it the client specifies
     * to send this using `ClientCapabilities.textDocument.completion.contextSupport === true`
     */
    context?: CompletionContext
}

/**
 * Request to request completion at a given text document position. The request's
 * parameter is of type [TextDocumentPosition](#TextDocumentPosition) the response
 * is of type [CompletionItem[]](#CompletionItem) or [CompletionList](#CompletionList)
 * or a Thenable that resolves to such.
 *
 * The request can delay the computation of the [`detail`](#CompletionItem.detail)
 * and [`documentation`](#CompletionItem.documentation) properties to the `completionItem/resolve`
 * request. However, properties that are needed for the initial sorting and filtering, like `sortText`,
 * `filterText`, `insertText`, and `textEdit`, must not be changed during resolve.
 */
export namespace CompletionRequest {
    export const type = new RequestType<
        CompletionParams,
        CompletionItem[] | CompletionList | null,
        void,
        CompletionRegistrationOptions
    >('textDocument/completion')
}

/**
 * Request to resolve additional information for a given completion item.The request's
 * parameter is of type [CompletionItem](#CompletionItem) the response
 * is of type [CompletionItem](#CompletionItem) or a Thenable that resolves to such.
 */
export namespace CompletionResolveRequest {
    export const type = new RequestType<CompletionItem, CompletionItem, void, void>('completionItem/resolve')
}
