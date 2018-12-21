import { AbortController } from 'abort-controller'
import { Observable, of } from 'rxjs'
import { bufferCount, delay } from 'rxjs/operators'
import { createBarrier } from '../../integration-test/testHelpers'
import { createConnection } from './connection'
import { ErrorCodes } from './messages'
import { createMessagePipe, createMessageTransports } from './testHelpers'

describe('Connection', () => {
    // Polyfill
    ;(global as any).AbortController = AbortController

    test('handle single request', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, (p1, _signal) => {
            expect(p1).toEqual(['foo'])
            return p1
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await expect(client.sendRequest(method, ['foo'])).resolves.toEqual(['foo'])
    })

    test('handle single request with async result', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, (p1, _signal) => {
            expect(p1).toEqual(['foo'])
            return Promise.resolve(p1)
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await expect(client.sendRequest(method, ['foo'])).resolves.toEqual(['foo'])
    })

    test('abort undispatched request', async () => {
        const [serverTransports, clientTransports] = createMessageTransports()
        const b1 = createBarrier()
        const b2 = createBarrier()

        const server = createConnection(serverTransports)
        server.onRequest('block', async () => {
            b2.done()
            await b1.wait
        })
        server.onRequest('undispatched', () => {
            throw new Error('handler should not be called')
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendRequest('block').catch(null)
        await b2.wait
        const abortController = new AbortController()
        const result = client.sendRequest('undispatched', ['foo'], abortController.signal)
        abortController.abort()
        b1.done()
        await expect(result).rejects.toHaveProperty('name', 'AbortError')
    })

    test('abort request currently being handled', async () => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest('m', (_params, signal) => {
            if (!signal) {
                throw new Error('!signal')
            }
            return new Promise<number>(resolve => {
                signal.addEventListener('abort', () => resolve(123))
            })
        })
        server.onRequest('ping', () => 'pong')
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        const abortController = new AbortController()
        const result = client.sendRequest('m', undefined, abortController.signal)
        await expect(client.sendRequest('ping')).resolves.toBe('pong') // waits until the 'm' message starts to be handled
        abortController.abort()
        await expect(result).rejects.toHaveProperty('name', 'AbortError')
    })

    test('send request with single observable emission', async () => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest('m', (params: [number]) => of(params[0] + 1).pipe(delay(0)))
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await expect(client.sendRequest<number>('m', [1])).resolves.toBe(2)
    })

    test('observe request with single observable emission', async () => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest('m', (params: [number]) => of(params[0] + 1))
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        const result = client.observeRequest<number>('m', [1])
        await expect(result.toPromise()).resolves.toEqual(2)
    })

    test('observe request with multiple observable emissions', async () => {
        const [serverTransports, clientTransports] = createMessageTransports()
        const server = createConnection(serverTransports)
        server.onRequest(
            'm',
            (params: number[]) =>
                new Observable<number>(observer => {
                    for (const v of params) {
                        observer.next(v + 1)
                    }
                    observer.complete()
                })
        )
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await expect(
            client
                .observeRequest<number>('m', [1, 2, 3, 4])
                .pipe(bufferCount(4))
                .toPromise()
        ).resolves.toEqual([2, 3, 4, 5])
    })

    test('handle multiple requests', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, (p1, _signal) => p1)
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        const promises: Promise<string>[] = []
        promises.push(client.sendRequest(method, ['foo']))
        promises.push(client.sendRequest(method, ['bar']))

        await expect(Promise.all(promises)).resolves.toEqual([['foo'], ['bar']])
    })

    test('unhandled request', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await expect(client.sendRequest(method, ['foo'])).rejects.toHaveProperty('code', ErrorCodes.MethodNotFound)
    })

    test('handler throws an Error', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, () => {
            throw new Error('test')
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        const result = client.sendRequest(method, ['foo'])
        await expect(result).rejects.toMatchObject({ code: ErrorCodes.InternalError, message: 'test' })
        await expect(result).rejects.toHaveProperty('data')
        await expect(result.catch(err => typeof err.data.stack)).resolves.toBe('string')
    })

    test('handler returns a rejected Promise with an Error', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, () => Promise.reject(new Error('test')))
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        const result = client.sendRequest(method, ['foo'])
        await expect(result).rejects.toMatchObject({ code: ErrorCodes.InternalError, message: 'test' })
        await expect(result).rejects.toHaveProperty('data')
        await expect(result.catch(err => typeof err.data.stack)).resolves.toBe('string')
    })

    test('receives undefined request params as null', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, params => {
            expect(params).toBe(null)
            return ''
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await client.sendRequest(method)
    })

    test('receives undefined notification params as null', async () => {
        const method = 'testNotification'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification(method, params => {
            expect(params).toBe(null)
            return ''
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification(method)
    })

    test('receives null as null', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, params => {
            expect(params).toEqual([null])
            return null
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await expect(client.sendRequest(method, [null])).resolves.toBe(null)
    })

    test('receives 0 as 0', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, params => {
            expect(params).toEqual([0])
            return 0
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await expect(client.sendRequest(method, [0])).resolves.toBe(0)
    })

    const testNotification = 'testNotification'
    test('sends and receives notification', done => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification(testNotification, params => {
            expect(params).toEqual([{ value: true }])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification(testNotification, [{ value: true }])
    })

    test('unhandled notification event', done => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onUnhandledNotification(message => {
            expect(message.method).toBe(testNotification)
            expect(message.params).toEqual([{ value: true }])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification(testNotification, [{ value: true }])
    })

    test('unsubscribes client connection', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const client = createConnection(clientTransports)
        const server = createConnection(serverTransports)
        server.onRequest(method, _param => {
            client.unsubscribe()
            return ''
        })
        server.listen()

        client.listen()
        await expect(client.sendRequest(method, [''])).rejects.toMatchObject({
            message:
                'The underlying JSON-RPC connection got unsubscribed while responding to this test/handleSingleRequest request.',
        })
        server.unsubscribe()
    })

    test('unsubscribed connection throws', () => {
        const client = createConnection(createMessagePipe())
        client.listen()
        client.unsubscribe()
        expect(() => client.sendNotification(testNotification)).toThrow()
    })

    test('two listen throw', () => {
        const client = createConnection(createMessagePipe())
        client.listen()
        expect(() => client.listen()).toThrow()
    })

    test('notify on connection unsubscribe', done => {
        const client = createConnection(createMessagePipe())
        client.listen()
        client.onUnsubscribe(() => {
            done()
        })
        client.unsubscribe()
    })

    test('params in notifications', done => {
        const method = 'test'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification(method, params => {
            expect(params).toEqual([10, 'vscode'])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification(method, [10, 'vscode'])
    })

    test('params in request/response', async () => {
        const method = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, (params: number[]) => {
            expect(params).toEqual([10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await expect(client.sendRequest(method, [10, 20, 30])).resolves.toBe(60)
    })

    test('params in request/response with signal', async () => {
        const method = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, (params: number[], _signal) => {
            expect(params).toEqual([10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        const signal = new AbortController().signal
        client.listen()
        await expect(client.sendRequest(method, [10, 20, 30], signal)).resolves.toBe(60)
    })

    test('1 param as array in request', async () => {
        const type = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(type, p1 => {
            expect(Array.isArray(p1)).toBeTruthy()
            expect(p1[0]).toBe(10)
            expect(p1[1]).toBe(20)
            expect(p1[2]).toBe(30)
            return 60
        })
        server.listen()

        const client = createConnection(clientTransports)
        const signal = new AbortController().signal
        client.listen()
        await expect(client.sendRequest(type, [10, 20, 30], signal)).resolves.toBe(60)
    })

    test('1 param as array in notification', done => {
        const type = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification(type, params => {
            expect(params).toEqual([10, 20, 30])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification(type, [10, 20, 30])
    })

    test('untyped request/response', async () => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest('test', (params: number[], _signal) => {
            expect(params).toEqual([10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        const signal = new AbortController().signal
        client.listen()
        await expect(client.sendRequest('test', [10, 20, 30], signal)).resolves.toBe(60)
    })

    test('untyped notification', done => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification('test', (params: number[]) => {
            expect(params).toEqual([10, 20, 30])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification('test', [10, 20, 30])
    })

    test('star request handler', async () => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest((method: string, params: number[], _signal) => {
            expect(method).toBe('test')
            expect(params).toEqual([10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        const signal = new AbortController().signal
        client.listen()
        await expect(client.sendRequest('test', [10, 20, 30], signal)).resolves.toBe(60)
    })

    test('star notification handler', done => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification((method: string, params: number[]) => {
            expect(method).toBe('test')
            expect(params).toEqual([10, 20, 30])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification('test', [10, 20, 30])
    })

    test('abort signal is undefined', async () => {
        const type = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(type, (params: number[], _signal) => {
            expect(params).toEqual([10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await expect(client.sendRequest(type, [10, 20, 30], undefined)).resolves.toBe(60)
    })

    test('null params in request', async () => {
        const type = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(type, _signal => 123)
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        ;(client.sendRequest as any)(type, null).then((result: any) => expect(result).toBe(123))
    })

    test('null params in notifications', done => {
        const type = 'test'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification(type, params => {
            expect(params).toBe(null)
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        ;(client.sendNotification as any)(type, null)
    })
})
