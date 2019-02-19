import * as MessageChannelAdapter from '@sourcegraph/comlink/messagechanneladapter'
import { Observable } from 'rxjs'
import uuid from 'uuid'
import { createExtensionHost as createInPageExtensionHost } from '../../../../shared/src/api/extension/worker'
import { EndpointPair } from '../../../../shared/src/platform/context'
import { isInPage } from '../context'

/**
 * Returns an observable of a communication channel to an extension host.
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

function endpointFromPort(port: chrome.runtime.Port): MessagePort {
    return MessageChannelAdapter.wrap({
        send: data => port.postMessage(data),
        addEventListener: (event, listener) => {
            port.onMessage.addListener(listener as any)
        },
        removeEventListener: (event, listener) => {
            port.onMessage.removeListener(listener as any)
        },
    })
}
