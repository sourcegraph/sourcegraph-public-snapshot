import { Endpoint } from 'comlink'
import { Observable, of } from 'rxjs'
import uuid from 'uuid'
import { EndpointPair } from '../../../../shared/src/platform/context'
import { isInPage } from '../context'
import { createExtensionHostWorker } from './worker'

/**
 * Spawns an extension and returns a communication channel to it.
 */
export function createExtensionHost(): Observable<EndpointPair> {
    if (isInPage) {
        return createInPageExtensionHost()
    }
    const id = uuid.v4()
    return of({
        proxy: endpointFromPort(
            chrome.runtime.connect({
                name: `proxy-${id}`,
            })
        ),
        expose: endpointFromPort(
            chrome.runtime.connect({
                name: `expose-${id}`,
            })
        ),
    })
}

export function endpointFromPort(port: chrome.runtime.Port): Endpoint {
    return {
        postMessage: data => port.postMessage({ data }),
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
    }
}

function createInPageExtensionHost(): Observable<EndpointPair> {
    throw new Error('Not implemented')
    // const worker = createExtensionHostWorker()
    // const messageTransports = createWebWorkerMessageTransports(worker)
    // return new Observable(sub => {
    //     sub.next(messageTransports)
    //     return () => worker.terminate()
    // })
}
