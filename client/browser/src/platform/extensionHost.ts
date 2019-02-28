import * as MessageChannelAdapter from '@sourcegraph/comlink/messagechanneladapter'
import { Observable } from 'rxjs'
import uuid from 'uuid'
import { createExtensionHost as createInPageExtensionHost } from '../../../../shared/src/api/extension/worker'
import { EndpointPair } from '../../../../shared/src/platform/context'
import { isInPage } from '../context'

/**
 * Returns an observable of a communication channel to an extension host.
 *
 * When executing in-page (for example as a Phabricator plugin), this simply
 * creates an extension host worker and emits the returned EndpointPair.
 *
 * When executing in the browser extension, we create pair of chrome.runtime.Port objects,
 * named 'expose-{uuid}' and 'proxy-{uuid}', and return the ports wrapped using ${@link endpointFromPort}.
 *
 * The background script will listen to newly created ports, create an extension host
 * worker per pair of ports, and forward messages between the port objects and
 * the extension host worker's endpoints.
 */
export function createExtensionHost(): Observable<EndpointPair> {
    if (isInPage) {
        return createInPageExtensionHost({ wrapEndpoints: false })
    }
    const id = uuid.v4()
    return new Observable(subscriber => {
        const proxyPort = chrome.runtime.connect({ name: `proxy-${id}` })
        const exposePort = chrome.runtime.connect({ name: `expose-${id}` })
        subscriber.next({
            proxy: endpointFromPort(proxyPort),
            expose: endpointFromPort(exposePort),
        })
        return () => {
            proxyPort.disconnect()
            exposePort.disconnect()
        }
    })
}

/**
 * Partially wraps a chrome.runtime.Port and returns a MessagePort created using
 * comlink's ${@link MessageChannelAdapter}, so that the Port can be used
 * as a comlink Endpoint to transport messages between the content script and the extension host.
 *
 * It is necessary to wrap the port using MessageChannelAdapter because chrome.runtime.Port objects do not support
 * transfering MessagePort objects (see https://github.com/GoogleChromeLabs/comlink/blob/master/messagechanneladapter.md).
 *
 */
function endpointFromPort(port: chrome.runtime.Port): MessagePort {
    const listeners = new Map<(event: MessageEvent) => any, (message: object, port: chrome.runtime.Port) => void>()
    return MessageChannelAdapter.wrap({
        send(data): void {
            port.postMessage(data)
        },
        addEventListener(event, messageListener): void {
            if (event !== 'message') {
                return
            }
            const chromePortListener = (data: object) => {
                messageListener.call(this, new MessageEvent('message', { data }))
            }
            listeners.set(messageListener, chromePortListener)
            port.onMessage.addListener(chromePortListener)
        },
        removeEventListener(event, messageListener): void {
            if (event !== 'message') {
                return
            }
            const chromePortListener = listeners.get(messageListener)
            if (!chromePortListener) {
                return
            }
            port.onMessage.removeListener(chromePortListener)
        },
    })
}
