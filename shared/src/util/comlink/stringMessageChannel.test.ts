import '../../api/integration-test/messagePortPolyfill' // TODO!(sqs): move this

import { StringMessagePort, wrapStringMessagePort } from './stringMessageChannel'

type HandlerFunction = (data: string) => void

class TestStringMessagePort implements StringMessagePort {
    public listeners: HandlerFunction[] = []

    public otherPort: TestStringMessagePort | undefined

    public dispatchMessage(data: string): void {
        for (const listener of this.listeners) {
            listener(data)
        }
    }

    public send(data: string): void {
        if (!this.otherPort) {
            throw new Error('no otherPort')
        }
        this.otherPort.dispatchMessage(data)
    }

    public addListener(listener: HandlerFunction): void {
        this.listeners.push(listener)
    }

    public removeListener(listener: HandlerFunction): void {
        const index = this.listeners.indexOf(listener)
        if (index !== -1) {
            this.listeners.splice(index, 1)
        }
    }
}

function createTestStringMessageChannel(): { port1: StringMessagePort; port2: StringMessagePort } {
    const port1 = new TestStringMessagePort()
    const port2 = new TestStringMessagePort()
    port1.otherPort = port2
    port2.otherPort = port1
    return { port1, port2 }
}

describe('createTestStringMessageChannel', () => {
    test('sends and receives', done => {
        const { port1, port2 } = createTestStringMessageChannel()
        port2.addListener(data => {
            expect(data).toBe('a')
            done()
        })
        port1.send('a')
    })
})

describe('wrapStringMessagePort', () => {
    const createWrappedStringMessageChannel = () => {
        const stringMessageChannel = createTestStringMessageChannel()
        return {
            port1: wrapStringMessagePort(stringMessageChannel.port1),
            port2: wrapStringMessagePort(stringMessageChannel.port2),
        }
    }

    test('sends and receives', done => {
        const { port1, port2 } = createWrappedStringMessageChannel()
        port2.addEventListener('message', ({ data }) => {
            expect(data).toBe('a')
            done()
        })
        port1.postMessage('a')
    })
})
