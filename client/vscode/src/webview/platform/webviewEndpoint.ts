import * as Comlink from 'comlink'
import { isObject } from 'lodash'

import { EndpointPair } from '@sourcegraph/shared/src/platform/context'

import { VsCodeApi } from '../vsCodeApi'

import { generateUUID } from './proxyTransferHandler'

const panelId = document.documentElement.dataset.panelId!

export function createEndpoints(vscodeApi: VsCodeApi): EndpointPair {
    const onMessages = new WeakMap<EventListenerOrEventListenerObject, EventListener>()

    const vscodeWebviewProxyTransferHandler: Comlink.TransferHandler<
        object,
        // Send panelId (in serialize), no need to receive it.
        { nestedConnectionId: string; panelId?: string }
    > = {
        canHandle: (value): value is Comlink.ProxyMarked =>
            isObject(value) && (value as Comlink.ProxyMarked)[Comlink.proxyMarker],
        serialize: proxyMarkedValue => {
            const nestedConnectionId = generateUUID()

            // Create endpoint, expose proxy marked value.
            const endpoint = createEndpoint(nestedConnectionId)
            Comlink.expose(proxyMarkedValue, endpoint)

            return [{ nestedConnectionId, panelId }, []]
        },
        deserialize: serialized => {
            // Create endpoint, return wrapped proxy.
            const endpoint = createEndpoint(serialized.nestedConnectionId)
            const proxy = Comlink.wrap(endpoint)

            return proxy
        },
    }

    Comlink.transferHandlers.set('proxy', vscodeWebviewProxyTransferHandler)

    function createEndpoint(connectionId: string): Comlink.Endpoint {
        return {
            postMessage: message => {
                vscodeApi.postMessage({ ...message, connectionId, panelId })
            },

            addEventListener: (type, listener) => {
                // This event listener will be called for all proxy method calls.
                // Comlink will send the message to the appropriate caller
                // based on UUID (generated internally).
                function onMessage(event: MessageEvent): void {
                    if (event.data?.connectionId === connectionId) {
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

    const extensionEndpoint = createEndpoint('extension')
    const webviewEndpoint = createEndpoint('webview')

    return {
        proxy: extensionEndpoint,
        expose: webviewEndpoint,
    }
}
