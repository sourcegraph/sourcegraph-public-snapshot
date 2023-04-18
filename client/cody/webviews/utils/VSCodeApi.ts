import { ExtensionMessage, WebviewMessage } from '../../src/chat/protocol'

declare const acquireVsCodeApi: () => VSCodeApi

interface VSCodeApi {
    getState: () => unknown
    setState: (newState: unknown) => unknown
    postMessage: (message: unknown) => void
}

class VSCodeWrapper {
    private readonly vscodeApi: VSCodeApi = acquireVsCodeApi()

    public postMessage(message: WebviewMessage): void {
        this.vscodeApi.postMessage(message)
    }

    public onMessage(callback: (message: ExtensionMessage) => void): () => void {
        const listener = (event: MessageEvent<ExtensionMessage>): void => {
            callback(event.data)
        }
        window.addEventListener('message', listener)
        return () => window.removeEventListener('message', listener)
    }
}

export const vscodeAPI: VSCodeWrapper = new VSCodeWrapper()
