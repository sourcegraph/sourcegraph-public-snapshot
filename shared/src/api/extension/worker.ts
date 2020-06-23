import ExtensionHostWorker from 'worker-loader?inline&name=extensionHostWorker.bundle.js!./main.worker.ts'
import { EndpointPair, ClosableEndpointPair } from '../../platform/context'
import { Subscription } from 'rxjs'

/**
 * Creates a web worker with the extension host and sets up a bidirectional MessageChannel-based communication channel.
 */
export function createExtensionHostWorker(): ClosableEndpointPair {
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
    const subscription = new Subscription(() => {
        extensionHostAPIChannel.port1.close()
        extensionHostAPIChannel.port2.close()
        clientAPIChannel.port1.close()
        clientAPIChannel.port2.close()
    })
    return { ...clientEndpoints, subscription }
}
