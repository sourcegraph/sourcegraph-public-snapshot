import { Endpoint } from '@sourcegraph/comlink'
import * as MessageChannelAdapter from '@sourcegraph/comlink/messagechanneladapter'
import { Observable, of } from 'rxjs'
import uuid from 'uuid'
import { EndpointPair } from '../../../../shared/src/platform/context'
import { isInPage } from '../context'
import { createExtensionHostWorker } from './worker'

/**
 * Returns an observable of a communication channel to an extension host.
 */
export function createExtensionHost(): Observable<EndpointPair> {
    if (isInPage) {
        return createInPageExtensionHost()
    }
    const id = uuid.v4()
    // TODO(lguychard): handle port disconnection
    return of({
        proxy: endpointFromPort(`proxy-${id}`),
        expose: endpointFromPort(`expose-${id}`),
    })
}

export function endpointFromPort(name: string): Endpoint {
    const port = chrome.runtime.connect({ name })
    return MessageChannelAdapter.wrap({
        send: data => port.postMessage(data),
        addEventListener: (event, listener) => {
            if (event !== 'message') {
                throw new Error(`Unhandled event: ${event}`)
            }
            port.onMessage.addListener(listener as any)
        },
        removeEventListener: (event, listener) => {
            if (event !== 'message') {
                throw new Error(`Unhandled event: ${event}`)
            }
            port.onMessage.removeListener(listener as any)
        },
    })
}

function createInPageExtensionHost(): Observable<EndpointPair> {
    // TODO(lguychard) fix copy pasta
    return new Observable(subscriber => {
        const worker = createExtensionHostWorker()
        const clientAPIChannel = new MessageChannel()
        const extensionHostAPIChannel = new MessageChannel()
        const workerEndpoints: EndpointPair = {
            proxy: clientAPIChannel.port2,
            expose: extensionHostAPIChannel.port2,
        }
        worker.postMessage({ endpoints: workerEndpoints, wrapEndpoints: false }, Object.values(workerEndpoints))
        const clientEndpoints: EndpointPair = {
            proxy: extensionHostAPIChannel.port1,
            expose: clientAPIChannel.port1,
        }
        subscriber.next(clientEndpoints)
        return () => worker.terminate()
    })
}
