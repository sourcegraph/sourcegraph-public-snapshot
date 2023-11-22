import type { EndpointPair } from '../../platform/context'

/**
 * Creates a web worker with the extension host and sets up a bidirectional MessageChannel-based communication channel.
 */
export function createExtensionHostWorker(workerBundleURL: string): { worker: Worker; clientEndpoints: EndpointPair } {
    const clientAPIChannel = new MessageChannel()
    const extensionHostAPIChannel = new MessageChannel()
    const worker = new Worker(workerBundleURL)
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
