import { switchMap } from 'abortable-rx'
import { Observable, Unsubscribable } from 'rxjs'
import { EndpointPair } from '../../platform/context'
import { InitData } from '../extension/extensionHost'
import { createExtensionHostClientConnection } from './connection'
import { Services } from './services'

export interface ExtensionHostClient extends Unsubscribable {
    /**
     * Closes the connection to the extension host and stops the controller from reestablishing new
     * connections.
     */
    unsubscribe(): void
}

/**
 * Creates a client to communicate with an extension host.
 *
 * @param extensionHostEndpoint An observable that emits the connection to the extension host each time a new
 * connection is established.
 */
export function createExtensionHostClient(
    services: Services,
    extensionHostEndpoint: Observable<EndpointPair>,
    initData: InitData
): ExtensionHostClient {
    const client = extensionHostEndpoint.pipe(
        switchMap(async (endpoints, _, signal) => {
            const client = await createExtensionHostClientConnection(endpoints, services, initData)
            signal.addEventListener('abort', () => client.unsubscribe(), { once: true })
        })
    )
    return client.subscribe()
}
