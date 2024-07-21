import type { EndpointPair } from '../../platform/context'

import { startExtensionHost } from './extensionHost'

/**
 * Creates a web worker with the extension host and sets up a bidirectional MessageChannel-based communication channel.
 */
export function createExtensionHostWorker(workerBundleURL: string): {
    terminate: () => void
    clientEndpoints: EndpointPair
} {
    const clientAPIChannel = new MessageChannel()
    const extensionHostAPIChannel = new MessageChannel()
    // TODO!(sqs): remove worker bundle from esbuild config

    // const worker = new Worker(workerBundleURL)
    const workerEndpoints: EndpointPair = {
        proxy: clientAPIChannel.port2,
        expose: extensionHostAPIChannel.port2,
    }
    const extensionHost = startExtensionHost(workerEndpoints)
    // worker.postMessage({ endpoints: workerEndpoints }, Object.values(workerEndpoints))

    const clientEndpoints = {
        proxy: extensionHostAPIChannel.port1,
        expose: clientAPIChannel.port1,
    }
    return { terminate: () => extensionHost.unsubscribe(), clientEndpoints }
}
