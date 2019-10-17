import { startExtensionHost } from '../../shared/src/api/extension/extensionHost'
import { Endpoint } from '@sourcegraph/comlink'
import { noop } from 'lodash'

const createEndpoint = (
    name: 'proxy' | 'expose'
): Endpoint & Pick<MessagePort, 'start'> & { dispatch: (data: any) => void } => {
    const listeners = new Set<(message: MessageEvent) => void>()
    return {
        dispatch: (data: unknown) => {
            console.log('dispatching to', listeners.size, 'listeners')
            for (const listener of listeners) {
                debugger
                listener({ data } as MessageEvent)
            }
        },
        start: noop,
        postMessage: message => {
            console.log('posting message to', name, message)
            self.postToPuppeteer({ recipient: name, data: message })
        },
        addEventListener: (type, listener) => {
            if (type !== 'message') {
                return
            }
            listeners.add(listener)
        },
        removeEventListener: (type, listener) => {
            if (type !== 'message') {
                return
            }
            listeners.delete(listener)
        },
    }
}

const endpoints = {
    proxy: createEndpoint('expose'),
    expose: createEndpoint('proxy'),
}
self.onPuppeteerMessage = ({ recipient, data }) => {
    console.log('dispatching message to', recipient, data)
    if (!['proxy', 'expose'].includes(recipient)) {
        throw new Error(`Invalid recipient "${recipient}"`)
    }
    endpoints[recipient].dispatch(data)
}

console.log('starting extension host')
startExtensionHost(endpoints)
