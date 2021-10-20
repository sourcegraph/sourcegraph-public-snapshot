import * as Comlink from 'comlink'

import { VsCodeApi } from '..'

/**
 * TODO explain (using `type` since we can't achieve bi-directional Comlink connections by transferring MessageChannels
 * like we do in the web app or browser extension)
 */
export function vsCodeWebviewEndpoint(vscodeApi: VsCodeApi, connectionType: 'webview' | 'extension'): Comlink.Endpoint {
    const onMessages = new WeakMap<EventListenerOrEventListenerObject, EventListener>()

    return {
        postMessage: message => vscodeApi.postMessage({ ...message, connectionType }),
        addEventListener: (type, listener) => {
            // Listener will be used as key for wrapped listener which filters for connectionType
            function onMessage(event: MessageEvent): void {
                if (event.data?.connectionType === connectionType) {
                    return typeof listener === 'function' ? listener(event) : listener.handleEvent(event)
                }
            }

            onMessages.set(listener, onMessage as EventListener)

            window.addEventListener('message', onMessage)
        },
        removeEventListener: (type, listener) => {
            const onMessage = onMessages.get(listener)

            window.removeEventListener('message', onMessage ?? listener)
        },
    }
}
