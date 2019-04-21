import '../../api/integration-test/messagePortPolyfill' // TODO!(sqs): move this

import * as comlink from '@sourcegraph/comlink'
import { Observable, of } from 'rxjs'
import { delay, switchMap, take, toArray } from 'rxjs/operators'
import { wrapRemoteObservable } from '../../api/client/api/common'
import { proxySubscribable, ProxySubscribable } from '../../api/extension/api/common'
import { createBarrier } from '../../api/integration-test/testHelpers'
import { StringMessagePort, wrapStringMessagePort } from './stringMessageChannel'

/////////////////////
//
/////////
jest.setTimeout(500)

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
                transferredPort: MessagePort
            }

            const { port1, port2 } = createWrappedStringMessageChannel()

            port2.addEventListener('message', event => {
                const { transferredPort }: Data = event.data
                transferredPort.addEventListener('message', ({ data }) => {
                    expect(data).toBe('a')
                    done()
                })
            })

            const transferChannel = new MessageChannel()
            port1.postMessage({ transferredPort: transferChannel.port1 }, [transferChannel.port1])
            transferChannel.port2.postMessage('a')
        })

        test('MessagePort roundtrip', done => {
            interface Data {
                transferredPort: MessagePort
            }

            const { port1, port2 } = createWrappedStringMessageChannel()

            port2.addEventListener('message', event => {
                const { transferredPort }: Data = event.data
                transferredPort.addEventListener('message', ({ data }) => {
                    expect(data).toBe('a')
                    transferredPort.postMessage('z')
                })
            })

            const transferChannel = new MessageChannel()
            port1.postMessage({ transferredPort: transferChannel.port1 }, [transferChannel.port1])
            transferChannel.port2.postMessage('a')
            transferChannel.port2.addEventListener('message', ({ data }) => {
                expect(data).toBe('z')
                done()
            })
        })
    })

    describe('garbage-collects listeners', () => {
        test('when the original holder closes the MessagePort', async () => {
            interface Data {
                transferredPort: MessagePort
            }

            const wrapper = createWrappedStringMessageChannel()
            expect(wrapper.port1ListenerCount).toBe(1)
            expect(wrapper.port2ListenerCount).toBe(1)

            const gotMessagePort = createBarrier()
            const gotMessage = createBarrier()

            wrapper.port2.addEventListener('message', event => {
                const { transferredPort }: Data = event.data
                transferredPort.addEventListener('message', ({ data }) => {
                    expect(data).toBe('a')
                    gotMessage.done()
                })
                gotMessagePort.done()
            })
            expect(wrapper.port1ListenerCount).toBe(1)
            expect(wrapper.port2ListenerCount).toBe(1)

            const transferChannel = new MessageChannel()
            wrapper.port1.postMessage({ transferredPort: transferChannel.port1 }, [transferChannel.port1])
            await gotMessagePort.wait
            expect(wrapper.port1ListenerCount).toBe(2)
            expect(wrapper.port2ListenerCount).toBe(2)
            // Both underlying ports have an additional listener now. The one that received the
            // MessagePort now has a listener for the multiplexed port. The one that sent the
            // MessagePort has a listener for the multiplexed port for when the MessagePort's transferee
            // uses it to send a message back.

            transferChannel.port2.postMessage('a')
            expect(wrapper.port1ListenerCount).toBe(2)
            expect(wrapper.port2ListenerCount).toBe(2)
            await gotMessage.wait

            transferChannel.port1.close()
            transferChannel.port2.close()
            expect(wrapper.port1ListenerCount).toBe(1)
            expect(wrapper.port2ListenerCount).toBe(1)
        })

        test('when the recipient closes the MessagePort', async () => {
            interface Data {
                transferredPort: MessagePort
            }

            const wrapper = createWrappedStringMessageChannel()
            expect(wrapper.port1ListenerCount).toBe(1)
            expect(wrapper.port2ListenerCount).toBe(1)

            const gotMessagePort = createBarrier()
            const gotMessage = createBarrier()

            wrapper.port2.addEventListener('message', event => {
                const { transferredPort }: Data = event.data
                transferredPort.addEventListener('message', ({ data }) => {
                    expect(data).toBe('a')
                    gotMessage.done()
                    transferredPort.close()
                })
                gotMessagePort.done()
            })
            expect(wrapper.port1ListenerCount).toBe(1)
            expect(wrapper.port2ListenerCount).toBe(1)

            const transferChannel = new MessageChannel()
            wrapper.port1.postMessage({ transferredPort: transferChannel.port1 }, [transferChannel.port1])
            await gotMessagePort.wait
            expect(wrapper.port1ListenerCount).toBe(2)
            expect(wrapper.port2ListenerCount).toBe(2)

            transferChannel.port2.postMessage('a')
            await gotMessage.wait

            // Run after `transferredPort.close()` has propagated.
            await new Promise(resolve => {
                setTimeout(() => {
                    expect(wrapper.port1ListenerCount).toBe(1)
                    expect(wrapper.port2ListenerCount).toBe(1)
                    resolve()
                })
            })
        })
    })

    describe('proxied Observables', () => {
        test.only('unsubscribes', async () => {
            let unsubscribed = 0
            let subscribed = 0
            const gotUnsubscribed = createBarrier()
            const observable = new Observable<number>(sub => {
                subscribed++
                sub.next(subscribed)
                console.log('SUB', subscribed)
                return () => {
                    unsubscribed++
                    gotUnsubscribed.done()
                    console.log('UNSUB', unsubscribed)
                }
            })

            const wrapper = createWrappedStringMessageChannel()

            comlink.expose(comlink.proxy(() => proxySubscribable(observable)), wrapper.port1)

            const remoteGetObservable = comlink.wrap<() => ProxySubscribable<number>>(wrapper.port2)
            const getObservable = () => wrapRemoteObservable<number>(remoteGetObservable())
            const getObservableLocal = () => observable

            // getObservable()
            //     .subscribe()
            //     .unsubscribe()
            //
            // const sub = getObservable().subscribe(x => {
            //     sub.unsubscribe()
            // })
            //
            // expect(
            //     await getObservable()
            //         .pipe(first())
            //         .toPromise()
            // ).toBe(1)
            //
            expect(
                await of(1)
                    .pipe(
                        switchMap(() => getObservable()),
                        take(1),
                        toArray()
                    )
                    .toPromise()
            ).toEqual([1])
            console.log('GOT VALUES')
            await gotUnsubscribed.wait
            expect(unsubscribed).toBe(1)
            expect(subscribed).toBe(1)
        })
    })
})
