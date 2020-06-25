import ExtensionHostWorker from 'worker-loader?name=extensionHostWorker.bundle.js&publicPath=js/!./main.worker.ts'
import { EndpointPair, ClosableEndpointPair } from '../../platform/context'
import { Subscription } from 'rxjs'

/**
 * Creates a web worker with the extension host and sets up a bidirectional MessageChannel-based communication channel.
 */
export function createExtensionHostWorker(): { worker: ExtensionHostWorker; clientEndpoints: EndpointPair } {
    const clientAPIChannel = new MessageChannel()
    const extensionHostAPIChannel = new MessageChannel()
    const worker = new ExtensionHostWorker()
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

export function createExtensionHost(): ClosableEndpointPair {
    const { clientEndpoints, worker } = createExtensionHostWorker()
    return { endpoints: clientEndpoints, subscription: new Subscription(() => worker.terminate()) }
}
