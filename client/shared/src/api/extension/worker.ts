// eslint-disable-next-line import/extensions
import { Subscription } from 'rxjs'

import { EndpointPair, ClosableEndpointPair } from '../../platform/context'

/* eslint-disable import/extensions, @typescript-eslint/ban-ts-comment */
// @ts-ignore
import ExtensionHostWorker from './main.worker'

/* eslint-enable import/extensions, @typescript-eslint/ban-ts-comment */

/**
 * Creates a web worker with the extension host and sets up a bidirectional MessageChannel-based communication channel.
 *
 * If a `workerBundleURL` is provided, it is used to create a new Worker(), instead of using the ExtensionHostWorker
 * returned by worker-loader. This is useful to load the worker bundle from a different path.
 */
export function createExtensionHostWorker(workerBundleURL?: string): { worker: Worker; clientEndpoints: EndpointPair } {
    const clientAPIChannel = new MessageChannel()
    const extensionHostAPIChannel = new MessageChannel()
    const worker = workerBundleURL ? new Worker(workerBundleURL) : new ExtensionHostWorker()
    const workerEndpoints: EndpointPair = {
        proxy: clientAPIChannel.port2,
        expose: extensionHostAPIChannel.port2,
    }
    worker.postMessage({ endpoints: workerEndpoints }, Object.values(workerEndpoints))
    const clientEndpoints = {
        proxy: extensionHostAPIChannel.port1,
        expose: clientAPIChannel.port1,
    }
    return { worker, clientEndpoints }
}

export function createExtensionHost(workerBundleURL?: string): ClosableEndpointPair {
    const { clientEndpoints, worker } = createExtensionHostWorker(workerBundleURL)
    return { endpoints: clientEndpoints, subscription: new Subscription(() => worker.terminate()) }
}
