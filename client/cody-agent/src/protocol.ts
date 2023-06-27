/* eslint-disable @typescript-eslint/consistent-type-definitions */
import { Completion } from '@sourcegraph/cody-shared/src/autocomplete'
import { RecipeID } from '@sourcegraph/cody-shared/src/chat/recipes/recipe'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

// This file documents the Cody Agent JSON-RPC protocol. Consult the JSON-RPC
// specification to learn about how JSON-RPC works https://www.jsonrpc.org/specification
// The Cody Agent server only supports transport via stdout/stdin.

// The JSON-RPC requests of the Cody Agent protocol. Requests are async
// functions that return some (possibly null) value.
export type Requests = {
    // ================
    // Client -> Server
    // ================

    // The 'initialize' request must be sent at the start of the connection
    // before any other request/notification is sent.
    initialize: [ClientInfo, ServerInfo]
    // The 'shutdown' request must be sent before terminating the agent process.
    shutdown: [null, null]

    // Client requests the agent server to lists all recipes that are supported
    // by the agent.
    'recipes/list': [null, RecipeInfo[]]
    // Client requests the agent server to execute an individual recipe.
    // The response is null because the AI/Assistant messages are streamed through
    // the chat/updateMessageInProgress notification. The flow to trigger a recipe
    // is like this:
    // client --- recipes/execute --> server
    // client <-- chat/updateMessageInProgress --- server
    //             ....
    // client <-- chat/updateMessageInProgress --- server
    'recipes/execute': [ExecuteRecipeParams, null]
    // Client requests manual completion
    'completions/manual': [ManualCompletionParams, Completion[] | null]

    // ================
    // Server -> Client
    // ================
}

// The JSON-RPC notifications of the Cody Agent protocol. Notifications are
// synchronous fire-and-forget messages that have no return value. Notifications are
// conventionally used to represent streams of values.
export type Notifications = {
    // ================
    // Client -> Server
    // ================

    // The 'initalized' notification must be sent after receiving the 'initialize' response.
    initialized: [null]
    // The 'exit' notification must be sent after the client receives the 'shutdown' response.
    exit: [null]

    // The server should use the provided connection configuration for all
    // subsequent requests/notications. The previous connection configuration
    // should no longer be used.
    'connectionConfiguration/didChange': [ConnectionConfiguration]

    // Lifecycle notifications for the client to notify the server about text
    // contents of documents and to notify which document is currently focused.
    'textDocument/didOpen': [TextDocument]
    // The 'textDocument/didChange' notification should be sent on almost every
    // keystroke, whether the text contents changed or the cursor/selection
    // changed.  Leave the `content` property undefined when the document's
    // content is unchanged.
    'textDocument/didChange': [TextDocument]
    // The user focused on a document without changing the document's content.
    // Only the 'uri' property is required, other properties are ignored.
    'textDocument/didFocus': [TextDocument]
    // The user closed the editor tab for the given document.
    // Only the 'uri' property is required, other properties are ignored.
    'textDocument/didClose': [TextDocument]

    // ================
    // Server -> Client
    // ================
    // The server received new messages for the ongoing 'chat/executeRecipe'
    // request. The server should never send this notification outside of a
    // 'chat/executeRecipe' request.
    'chat/updateMessageInProgress': [ChatMessage | null]
}

export interface ClientInfo {
    name: string
    version: string
    workspaceRootPath: string
    connectionConfiguration?: ConnectionConfiguration
    capabilities?: ClientCapabilities
}

export interface ClientCapabilities {
    completions?: 'none'
    //  When 'streaming', handles 'chat/updateMessageInProgress' streaming notifications.
    chat?: 'none' | 'streaming'
}

export interface ServerInfo {
    name: string
    capabilities?: ServerCapabilities
}
export interface ServerCapabilities {}

export interface ConnectionConfiguration {
    serverEndpoint: string
    accessToken: string
    customHeaders: Record<string, string>
}

export interface Position {
    // 0-indexed
    line: number
    // 0-indexed
    character: number
}

export interface Range {
    start: Position
    end: Position
}

// TODO: Add language id
export interface TextDocument {
    filePath: string
    content?: string
    selection?: Range
}

export interface RecipeInfo {
    id: RecipeID
    title: string
}

export interface ExecuteRecipeParams {
    id: RecipeID
    humanChatInput: string
}

export interface ManualCompletionParams {
    count: number
}
