import { Message, ResponseMessage } from '../jsonrpc2/messages'
import { isFunction } from '../util'

// Copied from vscode-jsonrpc to avoid adding extraneous dependencies.

export interface ConnectionStrategy {
    cancelUndispatched?: (
        message: Message,
        next: (message: Message) => ResponseMessage | undefined
    ) => ResponseMessage | undefined
}

export namespace ConnectionStrategy {
    export function is(value: any): value is ConnectionStrategy {
        const candidate: ConnectionStrategy = value
        return candidate && isFunction(candidate.cancelUndispatched)
    }
}
