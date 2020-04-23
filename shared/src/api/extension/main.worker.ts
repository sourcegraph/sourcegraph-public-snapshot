import '../../polyfills'

import * as MessageChannelAdapter from '@sourcegraph/comlink/dist/umd/string-channel.experimental'
import { fromEvent } from 'rxjs'
import { take } from 'rxjs/operators'
import { EndpointPair, isEndpointPair } from '../../platform/context'
import { startExtensionHost } from './extensionHost'
import { hasProperty } from '../../util/types'

interface InitMessage {
    endpoints: {
        proxy: MessagePort
        expose: MessagePort
    }
    /**
     * Whether the endpoints should be wrapped with a comlink {@link MessageChannelAdapter}.
     *
     * This is true when the messages passed on the endpoints are forwarded to/from
     * other wrapped endpoints, like in the browser extension.
     */
    wrapEndpoints: boolean
}

const isInitMessage = (value: unknown): value is InitMessage =>
    typeof value === 'object' && value !== null && hasProperty('endpoints')(value) && isEndpointPair(value.endpoints)

const wrapMessagePort = (port: MessagePort): MessagePort =>
    MessageChannelAdapter.wrap({
        send: data => port.postMessage(data),
        addMessageListener: listener => port.addEventListener('message', event => listener(event.data)),
    })

const wrapEndpoints = ({ proxy, expose }: InitMessage['endpoints']): EndpointPair => {
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
        const event = await fromEvent<MessageEvent>(self, 'message').pipe(take(1)).toPromise()
        if (!isInitMessage(event.data)) {
            throw new Error('First message event in extension host worker was not a well-formed InitMessage')
        }
        const { endpoints } = event.data
        const extensionHost = startExtensionHost(event.data.wrapEndpoints ? wrapEndpoints(endpoints) : endpoints)
        self.addEventListener('unload', () => extensionHost.unsubscribe())
    } catch (err) {
        console.error('Error starting the extension host:', err)
        self.close()
    }
}

// eslint-disable-next-line @typescript-eslint/no-floating-promises
extensionHostMain()
