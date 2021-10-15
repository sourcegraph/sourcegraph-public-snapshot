import * as Comlink from 'comlink'
import { isObject } from 'lodash'
import vscode from 'vscode'

import { EndpointPair } from '@sourcegraph/shared/src/platform/context'
import { hasProperty } from '@sourcegraph/shared/src/util/types'

import { generateUUID } from './proxyTransferHandler'

let nextPanelId = 1

function isProxyMarked(value: unknown): value is Comlink.ProxyMarked {
    return isObject(value) && (value as Comlink.ProxyMarked)[Comlink.proxyMarker]
}

interface NestedConnectionData {
    nestedConnectionId: string
    proxyMarkedValue?: object
    panelId: string
}

function isNestedConnection(value: unknown): value is NestedConnectionData {
    return isObject(value) && hasProperty('nestedConnectionId')(value) && hasProperty('proxyMarkedValue')(value)
}

// TODO explain, panelId -> factory, remove when panel is destroyed
const endpointFactories = new Map<string, ((connectionId: string) => Comlink.Endpoint) | undefined>()

// TODO clean up types, share w type guard
const vscodeExtensionProxyTransferHandler: Comlink.TransferHandler<
    object,
    // Receive panelId (in deserialize), no need to send it.
    // Return proxyMarkedValue in serialize so we can expose it in postMessage, but
    // only end up sending the nestedConnectionId
    { nestedConnectionId: string; proxyMarkedValue?: object; panelId?: string }
> = {
    canHandle: isProxyMarked,
    serialize: proxyMarkedValue => {
        const nestedConnectionId = generateUUID()
        // Defer endpoint creation/object exposition to `postMessage` (to scope it to panel)

        return [{ nestedConnectionId, proxyMarkedValue }, []]
    },
    deserialize: serialized => {
        // Create endpoint, return wrapped proxy.
        const endpointFactory = endpointFactories.get(serialized.panelId!)!
        const endpoint = endpointFactory(serialized.nestedConnectionId)
        const proxy = Comlink.wrap(endpoint)

        return proxy
    },
}

Comlink.transferHandlers.set('proxy', vscodeExtensionProxyTransferHandler)

export function createEndpointsForWebview(
    panel: Pick<vscode.WebviewPanel, 'onDidDispose' | 'webview'>
): EndpointPair & { panelId: string } {
    const listenerDisposables = new WeakMap<EventListenerOrEventListenerObject, vscode.Disposable>()
    const panelId = nextPanelId.toString()
    nextPanelId++
    let disposed = false

    function createEndpoint(connectionId: string): Comlink.Endpoint {
        /**
         * Handles values send to webviews that are marked to be proxied.
         */
        function toWireValue(value: NestedConnectionData): void {
            const proxyMarkedValue = value.proxyMarkedValue!
            // The proxyMarkedValue is probably not cloneable, so don't
            // send it "over the wire"
            delete value.proxyMarkedValue

            const endpoint = createEndpoint(value.nestedConnectionId)
            Comlink.expose(proxyMarkedValue, endpoint)
        }

        return {
            postMessage: (message: any) => {
                const value = message.value
                const argumentList = message.argumentList

                if (isNestedConnection(value)) {
                    toWireValue(value)
                }
                if (Array.isArray(argumentList)) {
                    for (const argument of argumentList) {
                        const value = argument.value
                        if (isNestedConnection(value)) {
                            toWireValue(value)
                        }
                    }
                }

                if (!disposed) {
                    panel.webview.postMessage({ ...message, connectionId, panelId }).then(
                        () => {},
                        error => console.error('postMessage error', error)
                    )
                }
            },
            addEventListener: (type, listener) => {
                // This event listener will be called for all proxy method calls.
                // Comlink will send the message to the appropriate caller
                // based on UUID (generated internally).
                function onMessage(message: any): void {
                    if (message?.connectionId === connectionId) {
                        // Comlink is listening for a message event, only uses the `data` property.
                        const messageEvent = {
                            data: message,
                        } as MessageEvent

                        return typeof listener === 'function'
                            ? listener(messageEvent)
                            : listener.handleEvent(messageEvent)
                    }
                }

                const disposable = panel.webview.onDidReceiveMessage(onMessage)
                listenerDisposables.set(listener, disposable)
            },
            removeEventListener: (type, listener) => {
                const disposable = listenerDisposables.get(listener)
                disposable?.dispose()
                listenerDisposables.delete(listener)
            },
        }
    }

    endpointFactories.set(panelId, createEndpoint)
    panel.onDidDispose(() => {
        disposed = true
        endpointFactories.delete(panelId)
        // TODO deregister all listeners
    })

    const webviewEndpoint = createEndpoint('webview')
    const extensionEndpoint = createEndpoint('extension')

    return {
        proxy: webviewEndpoint,
        expose: extensionEndpoint,
        panelId,
    }
}
