import * as MessageChannelAdapter from '@sourcegraph/comlink/messagechanneladapter'
import { Observable } from 'rxjs'
import uuid from 'uuid'
import { EndpointPair } from '../../../shared/src/platform/context'
import { isInPage } from '../context'
import { SourcegraphIntegrationURLs } from './context'

function createInPageExtensionHost({
    sourcegraphURL,
    assetsURL,
}: SourcegraphIntegrationURLs): Observable<EndpointPair> {
    return new Observable(subscriber => {
        // Create an iframe pointing to extensionHostFrame.html,
        // which will load the extension host worker, and forward it
        // the client endpoints.
        const frame: HTMLIFrameElement = document.createElement('iframe')
        frame.setAttribute('src', new URL('extensionHostFrame.html', assetsURL).href)
        frame.setAttribute('style', 'display: none;')
        document.body.append(frame)
        const clientAPIChannel = new MessageChannel()
        const extensionHostAPIChannel = new MessageChannel()
        const workerEndpoints: EndpointPair = {
            proxy: clientAPIChannel.port2,
            expose: extensionHostAPIChannel.port2,
        }
        const clientEndpoints = {
            proxy: extensionHostAPIChannel.port1,
            expose: clientAPIChannel.port1,
        }
        // Subscribe to the load event on the frame,
        frame.addEventListener(
            'load',
            () => {
                frame.contentWindow!.postMessage(
                    {
                        type: 'workerInit',
                        payload: {
                            endpoints: clientEndpoints,
                            wrapEndpoints: false,
                        },
                    },
                    sourcegraphURL,
                    Object.values(clientEndpoints)
                )
                subscriber.next(workerEndpoints)
            },
            {
                once: true,
            }
        )
        return () => {
            clientEndpoints.proxy.close()
            clientEndpoints.expose.close()
            frame.remove()
        }
    })
}

/**
 * Returns an observable of a communication channel to an extension host.
 *
 * When executing in-page (for example as a Phabricator plugin), this simply
 * creates an extension host worker and emits the returned EndpointPair.
 *
 * When executing in the browser extension, we create pair of browser.runtime.Port objects,
 * named 'expose-{uuid}' and 'proxy-{uuid}', and return the ports wrapped using ${@link endpointFromPort}.
 *
 * The background script will listen to newly created ports, create an extension host
 * worker per pair of ports, and forward messages between the port objects and
 * the extension host worker's endpoints.
 */
export function createExtensionHost(urls: SourcegraphIntegrationURLs): Observable<EndpointPair> {
    if (isInPage) {
        return createInPageExtensionHost(urls)
    }
    const id = uuid.v4()
    return new Observable(subscriber => {
        const proxyPort = browser.runtime.connect({ name: `proxy-${id}` })
        const exposePort = browser.runtime.connect({ name: `expose-${id}` })
        subscriber.next({
            proxy: endpointFromPort(proxyPort),
            expose: endpointFromPort(exposePort),
        })
        return () => {
            proxyPort.disconnect()
            exposePort.disconnect()
        }
    })
}

/**
 * Partially wraps a browser.runtime.Port and returns a MessagePort created using
 * comlink's {@link MessageChannelAdapter}, so that the Port can be used
 * as a comlink Endpoint to transport messages between the content script and the extension host.
 *
 * It is necessary to wrap the port using MessageChannelAdapter because browser.runtime.Port objects do not support
 * transfering MessagePort objects (see https://github.com/GoogleChromeLabs/comlink/blob/master/messagechanneladapter.md).
 *
 */
function endpointFromPort(port: browser.runtime.Port): MessagePort {
    const messageListeners = new Map<(event: MessageEvent) => any, (message: unknown) => void>()
    return MessageChannelAdapter.wrap({
        send(data: string): void {
            port.postMessage(data)
        },
        addEventListener(event: 'message', messageListener: (event: MessageEvent) => any): void {
            if (event !== 'message') {
                return
            }
            const portListener = (data: unknown): void => {
                // This callback is called *very* often (e.g., ~900 times per keystroke in a
                // monitored textarea). Avoid creating unneeded objects here because GC
                // significantly hurts perf. See
                // https://github.com/sourcegraph/sourcegraph/issues/3433#issuecomment-483561297 and
                // watch that issue for a (possibly better) fix.
                //
                // HACK: Use a simple object here instead of `new MessageEvent('message', { data })`
                // to reduce the amount of garbage created. There are no callers that depend on
                // other MessageEvent properties; they would be set to their default values anyway,
                // so losing the properties is not a big problem.
                messageListener.call(this, { data } as MessageEvent)
            }
            messageListeners.set(messageListener, portListener)
            port.onMessage.addListener(portListener)
        },
        removeEventListener(event: 'message', messageListener: (event: MessageEvent) => any): void {
            if (event !== 'message') {
                return
            }
            const portListener = messageListeners.get(messageListener)
            if (!portListener) {
                return
            }
            port.onMessage.removeListener(portListener)
        },
    })
}
