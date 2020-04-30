import { isObject } from 'lodash'
import * as uuid from 'uuid'

const PORT_PREFIX = 'comlink-'

/**
 * Partially wraps a browser.runtime.Port and returns a MessagePort created using
 * comlink's {@link MessageChannelAdapter}, so that the Port can be used
 * as a comlink Endpoint to transport messages between the content script and the extension host.
 *
 * It is necessary to wrap the port using MessageChannelAdapter because browser.runtime.Port objects do not support
 * transferring MessagePort objects (see https://github.com/GoogleChromeLabs/comlink/blob/master/messagechanneladapter.md).
 */
export function browserPortToMessagePort(
    browserPort: browser.runtime.Port,
    connect: (name: string) => browser.runtime.Port
): MessagePort {
    const { port1: returned, port2: adapterPort } = new MessageChannel()

    const connectedBrowserPorts = new Map<string, browser.runtime.Port>()
    const waitingForPorts = new Map<string, (port: browser.runtime.Port) => void>()
    browser.runtime.onConnect.addListener(port => {
        console.log('received browser conection', port.name)
        const id = port.name.slice(PORT_PREFIX.length)
        const waitingFn = waitingForPorts.get(id)
        if (waitingFn) {
            console.log('was waiting', id)
            waitingFn(port)
        } else {
            console.log('noone waiting, saving', id)
            connectedBrowserPorts.set(id, port)
        }
    })
    /** Run the given callback as soon as the given port ID is connected. */
    const whenConnected = (id: string, cb: (port: browser.runtime.Port) => void): void => {
        const browserPort = connectedBrowserPorts.get(id)
        if (browserPort) {
            console.log('browser port available for', id)
            cb(browserPort)
            return
        }
        console.log('waiting for browser port', id)
        waitingForPorts.set(id, cb)
    }

    function link(browserPort: browser.runtime.Port, adapterPort: MessagePort): void {
        adapterPort.addEventListener('message', event => {
            let data: unknown = event.data
            // Message from comlink needing to be forwarded to browser port with MessagePorts removed
            const portRefs: PortRef[] = []
            // Find message port references and connect to the other side with the found IDs.
            for (const { value, path, key, parent } of iterateAllProperties(data)) {
                if (value instanceof MessagePort) {
                    // Remove MessagePort from message
                    if (key !== null) {
                        parent[key] = null
                    } else {
                        data = null
                    }
                    // Open browser port in place of MessagePort
                    const id = uuid.v4()
                    const browserPort = connect(PORT_PREFIX + id)
                    link(browserPort, value)
                    // Include the ID of the browser port in the message
                    portRefs.push({ path, id })
                }
            }
            // Wrap message for the browser port to include all port IDs
            const browserPortMessage: BrowserPortMessage = { message: data, portRefs }
            console.log('posting to browser port', browserPort.name, browserPortMessage)
            browserPort.postMessage(browserPortMessage)
        })

        browserPort.onMessage.addListener(({ message, portRefs }: BrowserPortMessage): void => {
            console.log('received on browser port', browserPort.name, { message, portRefs })
            const transfer: MessagePort[] = []
            for (const portRef of portRefs) {
                const { port1: comlinkMessagePort, port2: intermediateMessagePort } = new MessageChannel()
                // Add back MessagePorts at the port references that get hooked up to a browser Port
                replaceValueAtPath(message, portRef.path, comlinkMessagePort)
                // Need to mention them in the transfer parameter too
                transfer.push(comlinkMessagePort)
                // Once the port with the mentioned ID is connected, link it up
                whenConnected(portRef.id, browserPort => link(browserPort, intermediateMessagePort))
            }
            // Forward message, with MessagePorts
            console.log('posting to adapter port', message)
            adapterPort.postMessage(message, transfer)
        })

        browserPort.onDisconnect.addListener(() => {
            console.log('port disconnected, closing MessagePort', browserPort.name)
            adapterPort.close()
        })

        adapterPort.start()
    }
    link(browserPort, adapterPort)

    return returned
}

interface PortRef {
    /** Path at which the MessagePort appeared. */
    path: Path

    /** UUID of the browser Port that was created for it. */
    id: string
}

interface BrowserPortMessage {
    message: unknown
    portRefs: PortRef[]
}

type Key = string | number | symbol
type Path = Key[]

interface PropertyIteratorEntry<T = unknown> {
    value: T
    path: Path
    key: Key | null
    parent: any | null
}

function replaceValueAtPath(value: any, path: Path, newValue: unknown): unknown {
    const lastProp = path[path.length - 1]
    for (const prop of path.slice(0, -1)) {
        value = value[prop]
    }
    const oldValue = value[lastProp]
    value[lastProp] = newValue
    return oldValue
}

export function* iterateAllProperties<T>(
    value: T,
    key: Key | null = null,
    parent: any = null,
    path: Path = [],
    visited = new WeakSet()
): Iterable<PropertyIteratorEntry> {
    yield { value, path, parent, key }
    if (!isObject(value) || visited.has(value)) {
        return
    }
    visited.add(value)

    const keys = Object.keys(value) as (keyof typeof value)[]
    for (const key of keys) {
        yield* iterateAllProperties(value[key], key, value, [...path, key], visited)
    }
}

export function* findAllTransferables(message: unknown): Iterable<Transferable> {
    for (const { value } of iterateAllProperties(message)) {
        if (value instanceof MessagePort) {
            yield value
        }
    }
}
