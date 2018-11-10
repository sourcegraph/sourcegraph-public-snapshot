import { Message, ResponseMessage } from './messages'

// Copied from vscode-jsonrpc to avoid adding extraneous dependencies.

export interface ConnectionStrategy {
    cancelUndispatched?: (
        message: Message,
        next: (message: Message) => ResponseMessage | undefined
    ) => ResponseMessage | undefined
}
