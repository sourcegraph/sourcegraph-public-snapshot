import { RecipeID } from '@sourcegraph/cody-shared/src/chat/recipes/recipe'
import { TranscriptJSON } from '@sourcegraph/cody-shared/src/chat/transcript'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

// The RPC is packaged in the same way as LSP:
// Content-Length: lengthInBytes\r\n
// \r\n
// ...

// The RPC initialization process is the same as LSP:
// (-- Server process started; session begins --)
// Client: initialize request
// Server: initialize response
// Client: initialized notification
// Client and server send anything they want after this point
// The RPC shutdown process is the same as LSP:
// Client: shutdown request
// Server: shutdown response
// Client: exit notification
// (-- Server process exited; session ends --)
// eslint-disable-next-line @typescript-eslint/consistent-type-definitions
export type Requests = {
    // Client -> Server
    initialize: [ClientInfo, ServerInfo]
    shutdown: [void, void]

    'recipes/list': [void, RecipeInfo[]]
    'recipes/execute': [ExecuteRecipeParams, void]

    // Server -> Client
}

// eslint-disable-next-line @typescript-eslint/consistent-type-definitions
export type Notifications = {
    // Client -> Server
    initialized: [void]
    exit: [void]

    'workspaceRootPath/didChange': [string]
    'configuration/didChange': [Configuration]
    'textDocument/didFocus': [TextDocument]
    'textDocument/didOpen': [TextDocument]
    'textDocument/didChange': [TextDocument]
    'textDocument/didClose': [TextDocument]

    // Server -> Client
    'chat/updateMessageInProgress': [ChatMessage | null]
    'chat/updateTranscript': [TranscriptJSON]
}

export interface Configuration {
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

export interface TextDocument {
    filePath: string
    content?: string
    selection?: Range
}

export interface RecipeInfo {
    id: RecipeID
    title: string
}

export interface StaticEditor {
    workspaceRoot: string | null
}

// Static recipe context that lots of recipes might need
// More context is obtained if necessary via server to client requests
export interface StaticRecipeContext {
    editor: StaticEditor
}

export interface ExecuteRecipeParams {
    id: RecipeID
    humanChatInput: string
    context: StaticRecipeContext
}

export interface ReplaceSelectionParams {
    fileName: string
    selectedText: string
    replacement: string
}

export interface ReplaceSelectionResult {
    applied: boolean
    failureReason: string
}

// TODO: Add some version info to prevent version incompatibilities
// TODO: Add capabilities so clients can announce what they can handle
export interface ClientInfo {
    name: string
}

export interface ServerInfo {
    name: string
}
