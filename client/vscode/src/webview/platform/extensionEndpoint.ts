import * as Comlink from 'comlink'
import vscode from 'vscode'

import { RuntimeConnectionType } from './connection'

/**
 * TODO explain
 *
 * Note: this should only be used in the VS Code extension host runtime.
 */
export function vscodeExtensionEndpoint(
    panel: vscode.WebviewPanel,
    connectionType: RuntimeConnectionType
): Comlink.Endpoint {
    // TODO return endpoint and disposable fn to add to top level disposable?
    const listenerDisposables = new WeakMap<EventListenerOrEventListenerObject, vscode.Disposable>()

    return {
        postMessage: (message: any) => panel.webview.postMessage({ ...message, connectionType }),
        addEventListener: (type, listener) => {
            function onMessage(message: any): void {
                if (message?.connectionType === connectionType) {
                    // Comlink is listening for a message event, only uses the `data` property.
                    const messageEvent = {
                        data: message,
                    } as MessageEvent

                    return typeof listener === 'function' ? listener(messageEvent) : listener.handleEvent(messageEvent)
                }
            }

            const disposable = panel.webview.onDidReceiveMessage(onMessage)
            listenerDisposables.set(listener, disposable)
        },
        removeEventListener: (type, listener) => {
            listenerDisposables.delete(listener)
        },
    }
}
