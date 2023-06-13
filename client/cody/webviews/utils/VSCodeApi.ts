import { ExtensionMessage, WebviewMessage } from '../../src/chat/protocol'

declare const acquireVsCodeApi: () => VSCodeApi

interface VSCodeApi {
    getState: () => unknown
    setState: (newState: unknown) => unknown
    postMessage: (message: unknown) => void
}

export interface VSCodeWrapper {
    postMessage(message: WebviewMessage): void
    onMessage(callback: (message: ExtensionMessage) => void): () => void
}

let api: VSCodeWrapper

export function getVSCodeAPI(): VSCodeWrapper {
    if (!api) {
        const vsCodeApi = acquireVsCodeApi()
        api = {
            postMessage: message => vsCodeApi.postMessage(message),
            onMessage: callback => {
                const listener = (event: MessageEvent<ExtensionMessage>): void => {
                    callback(event.data)
                }
                window.addEventListener('message', listener)
                return () => window.removeEventListener('message', listener)
            },
        }
    }
    return api
}
