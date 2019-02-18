import '../../polyfills'

import { fromEvent } from 'rxjs'
import { take } from 'rxjs/operators'
import { isEndpointPair } from '../../platform/context'
import { startExtensionHost } from './extensionHost'

/**
 * The entrypoint for the JavaScript context that runs the extension host (and all extensions).
 *
 * To initialize the extension host, the parent sends it an "initialize" message with
 * {@link InitData}.
 */
async function extensionHostMain(): Promise<void> {
    try {
        const event = await fromEvent<MessageEvent>(self, 'message')
            .pipe(take(1))
            .toPromise()
        if (!isEndpointPair(event.data)) {
            throw new Error('First message event in extension host worker did not contain MessagePort')
        }
        const endpoints = event.data
        endpoints.proxy.addEventListener('message', event =>
            console.log('Extension host received message on proxy port', event.data)
        )
        endpoints.expose.addEventListener('message', event =>
            console.log('Extension host received message on expose port', event.data)
        )
        const extensionHost = startExtensionHost(endpoints)
        self.addEventListener('unload', () => extensionHost.unsubscribe())
    } catch (err) {
        console.error('Error starting the extension host:', err)
        self.close()
    }
}

// tslint:disable-next-line: no-floating-promises
extensionHostMain()
