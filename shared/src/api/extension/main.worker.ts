import '../../polyfills'

import { Endpoint } from '@sourcegraph/comlink'
import * as MessageChannelAdapter from '@sourcegraph/comlink/messagechanneladapter'
import { fromEvent } from 'rxjs'
import { take } from 'rxjs/operators'
import { EndpointPair, isEndpointPair } from '../../platform/context'
import { startExtensionHost } from './extensionHost'

export interface InitMessage {
    endpoints: {
        proxy: MessagePort
        expose: MessagePort
    }
    /**
     * Whether the endpoints should be wrapped with a {@link MessageChannelAdapter}.
     *
     * This is true when the messages passed on the endpoints are forwarded to/from
     * other wrapped endpoints, like in the browser extension
     */
    wrapEndpoints: boolean
}

const isInitMessage = (m: any): m is InitMessage => m.endpoints && isEndpointPair(m.endpoints)

const wrapMessagePort = (port: MessagePort): Endpoint =>
    MessageChannelAdapter.wrap({
        send: data => port.postMessage(data),
        addEventListener: (event, listener) => port.addEventListener(event, listener),
        removeEventListener: (event, listener) => port.removeEventListener(event, listener),
    })

const wrappedEndpoints = ({ proxy, expose }: InitMessage['endpoints']): EndpointPair => {
    proxy.start()
    expose.start()
    return {
        proxy: wrapMessagePort(proxy),
        expose: wrapMessagePort(expose),
    }
}

/**
 * The entrypoint for the JavaScript context that runs the extension host (and all extensions).
 *
 * To initialize the extension host, the parent sends it an {@link InitMessage}
 */
async function extensionHostMain(): Promise<void> {
    try {
        const event = await fromEvent<MessageEvent>(self, 'message')
            .pipe(take(1))
            .toPromise()
        if (!isInitMessage(event.data)) {
            throw new Error('First message event in extension host worker was not a well-formed InitMessage')
        }
        const { endpoints, wrapEndpoints } = event.data
        // TODO support traceExtensionHostCommunication
        endpoints.proxy.addEventListener('message', event =>
            console.log('Extension host received message on proxy port', event.data)
        )
        endpoints.expose.addEventListener('message', event =>
            console.log('Extension host received message on expose port', event.data)
        )
        const extensionHost = startExtensionHost(wrapEndpoints ? wrappedEndpoints(endpoints) : endpoints)
        self.addEventListener('unload', () => extensionHost.unsubscribe())
    } catch (err) {
        console.error('Error starting the extension host:', err)
        self.close()
    }
}

// tslint:disable-next-line: no-floating-promises
extensionHostMain()
