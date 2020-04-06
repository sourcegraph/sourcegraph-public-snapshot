import { Observable } from 'rxjs'
import ExtensionHostWorker from 'worker-loader?inline&name=extensionHostWorker.bundle.js!./main.worker.ts'
import { EndpointPair } from '../../platform/context'

interface ExtensionHostInitOptions {
    /**
     * Whether the endpoints should be wrapped with a comlink {@link MessageChannelAdapter}.
     *
     * This is true when the messages passed on the endpoints are forwarded to/from
     * other wrapped endpoints, like in the browser extension.
     */
    wrapEndpoints: boolean
}

export function createExtensionHostWorker({
    wrapEndpoints,
}: ExtensionHostInitOptions): { worker: ExtensionHostWorker; clientEndpoints: EndpointPair } {
    const clientAPIChannel = new MessageChannel()
    const extensionHostAPIChannel = new MessageChannel()
    const worker = new ExtensionHostWorker()
    const workerEndpoints: EndpointPair = {
        proxy: clientAPIChannel.port2,
        expose: extensionHostAPIChannel.port2,
    }
    worker.postMessage({ endpoints: workerEndpoints, wrapEndpoints }, Object.values(workerEndpoints))
    const clientEndpoints = {
        proxy: extensionHostAPIChannel.port1,
        expose: clientAPIChannel.port1,
    }
    return { worker, clientEndpoints }
}

export function createExtensionHost({ wrapEndpoints }: ExtensionHostInitOptions): Observable<EndpointPair> {
    return new Observable(subscriber => {
        const { clientEndpoints, worker } = createExtensionHostWorker({ wrapEndpoints })
        subscriber.next(clientEndpoints)
        return () => worker.terminate()
    })
}
