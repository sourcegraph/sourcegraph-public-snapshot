import '../../api/integration-test/messagePortPolyfill' // TODO!(sqs): move this

import { createBarrier } from '../../api/integration-test/testHelpers'
import { StringMessagePort, wrapStringMessagePort } from './stringMessageChannel'

/////////////////////
//
/////////
jest.setTimeout(400)

type HandlerFunction = (data: string) => void

class TestStringMessagePort implements StringMessagePort {
    private listeners: HandlerFunction[] = []

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

    public get listenerCount(): number {
        return this.listeners.length
    }
}

function createTestStringMessageChannel(): { port1: TestStringMessagePort; port2: TestStringMessagePort } {
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
            get port1ListenerCount(): number {
                return stringMessageChannel.port1.listenerCount
            },
            port2: wrapStringMessagePort(stringMessageChannel.port2),
            get port2ListenerCount(): number {
                return stringMessageChannel.port2.listenerCount
            },
        }
    }

    describe('sends and receives', () => {
        test('primitive values', done => {
            const { port1, port2 } = createWrappedStringMessageChannel()
            port2.addEventListener('message', ({ data }) => {
                expect(data).toBe('a')
                done()
            })
            port1.postMessage('a')
        })

        test('MessagePort one-way', done => {
            interface Data {
                transferredChannel: MessagePort
            }

            const { port1, port2 } = createWrappedStringMessageChannel()

            port2.addEventListener('message', event => {
                const { transferredChannel }: Data = event.data
                transferredChannel.addEventListener('message', ({ data }) => {
                    expect(data).toBe('a')
                    done()
                })
            })

            const transferChannel = new MessageChannel()
            port1.postMessage({ transferredChannel: transferChannel.port1 }, [transferChannel.port1])
            transferChannel.port2.postMessage('a')
        })

        test('MessagePort roundtrip', done => {
            interface Data {
                transferredChannel: MessagePort
            }

            const { port1, port2 } = createWrappedStringMessageChannel()

            port2.addEventListener('message', event => {
                const { transferredChannel }: Data = event.data
                transferredChannel.addEventListener('message', ({ data }) => {
                    expect(data).toBe('a')
                    transferredChannel.postMessage('z')
                })
            })

            const transferChannel = new MessageChannel()
            port1.postMessage({ transferredChannel: transferChannel.port1 }, [transferChannel.port1])
            transferChannel.port2.postMessage('a')
            transferChannel.port2.addEventListener('message', ({ data }) => {
                expect(data).toBe('z')
                done()
            })
        })
    })

    test('garbage-collects listeners', async () => {
        interface Data {
            abc: MessagePort
        }

        const wrapper = createWrappedStringMessageChannel()
        expect(wrapper.port1ListenerCount).toBe(1)
        expect(wrapper.port2ListenerCount).toBe(1)

        const gotMessage = createBarrier()

        let receivedMessage = false
        wrapper.port2.addEventListener('message', event => {
            const data: Data = event.data
            data.abc.addEventListener('message', ({ data }) => {
                expect(data).toBe('a')
                receivedMessage = true
            })
            gotMessage.done()
        })
        expect(wrapper.port1ListenerCount).toBe(1)
        expect(wrapper.port2ListenerCount).toBe(1)

        const transferChannel = new MessageChannel()
        wrapper.port1.postMessage({ abc: transferChannel.port1 }, [transferChannel.port1])
        await gotMessage.wait
        expect(wrapper.port1ListenerCount).toBe(2)
        expect(wrapper.port2ListenerCount).toBe(2)
        // Both underlying ports have an additional listener now. The one that received the
        // MessagePort now has a listener for the multiplexed port. The one that sent the
        // MessagePort has a listener for the multiplexed port for when the MessagePort's transferee
        // uses it to send a message back.

        transferChannel.port2.postMessage('a')
        expect(wrapper.port1ListenerCount).toBe(2)
        expect(wrapper.port2ListenerCount).toBe(2)

        transferChannel.port1.close()
        transferChannel.port2.close()
        expect(wrapper.port1ListenerCount).toBe(1)
        expect(wrapper.port2ListenerCount).toBe(1)

        expect(receivedMessage).toBeTruthy()
    })

    test('asdf', () => {
        const mc = new MessageChannel()
        mc.port2.onmessageerror = ev => {
            console.error('ERROR', ev)
        }
        mc.port1.close()
        expect(1).toBe(2)
    })
})
