import { AbortController } from 'abort-controller'
import assert from 'assert'
import { createConnection } from './connection'
import { createMessagePipe, createMessageTransports } from './helpers.test'
import { ErrorCodes, ResponseError } from './messages'

describe('Connection', () => {
    // Polyfill
    ;(global as any).AbortController = AbortController

    it('handle single request', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, (p1, _signal) => {
            assert.deepStrictEqual(p1, ['foo'])
            return p1
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        assert.deepStrictEqual(await client.sendRequest(method, ['foo']), ['foo'])
    })

    it('handle single request with async result', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, (p1, _signal) => {
            assert.deepStrictEqual(p1, ['foo'])
            return Promise.resolve(p1)
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        assert.deepStrictEqual(await client.sendRequest(method, ['foo']), ['foo'])
    })

    it('handle multiple requests', async () => {
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

        const values = await Promise.all(promises)
        assert.deepStrictEqual(values, [['foo'], ['bar']])
    })

    it('unhandled request', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await assert.rejects(
            () => client.sendRequest(method, ['foo']),
            (error: ResponseError<any>) => error.code === ErrorCodes.MethodNotFound
        )
    })

    it('handler throws an Error', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, () => {
            throw new Error('test')
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await assert.rejects(
            () => client.sendRequest(method, ['foo']),
            (error: ResponseError<any>) =>
                error.code === ErrorCodes.InternalError &&
                error.message === 'test' &&
                error.data &&
                typeof error.data.stack === 'string'
        )
    })

    it('handler returns a rejected Promise with an Error', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, () => Promise.reject(new Error('test')))
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await assert.rejects(
            () => client.sendRequest(method, ['foo']),
            (error: ResponseError<any>) =>
                error.code === ErrorCodes.InternalError &&
                error.message === 'test' &&
                error.data &&
                typeof error.data.stack === 'string'
        )
    })

    it('receives undefined request params as null', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, params => {
            assert.strictEqual(params, null)
            return ''
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        await client.sendRequest(method)
    })

    it('receives undefined notification params as null', async () => {
        const method = 'testNotification'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification(method, params => {
            assert.strictEqual(params, null)
            return ''
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification(method)
    })

    it('receives null as null', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, params => {
            assert.deepStrictEqual(params, [null])
            return null
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        assert.strictEqual(await client.sendRequest(method, [null]), null)
    })

    it('receives 0 as 0', async () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, params => {
            assert.deepStrictEqual(params, [0])
            return 0
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        assert.strictEqual(await client.sendRequest(method, [0]), 0)
    })

    const testNotification = 'testNotification'
    it('sends and receives notification', done => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification(testNotification, params => {
            assert.deepStrictEqual(params, [{ value: true }])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification(testNotification, [{ value: true }])
    })

    it('unhandled notification event', done => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onUnhandledNotification(message => {
            assert.strictEqual(message.method, testNotification)
            assert.deepStrictEqual(message.params, [{ value: true }])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification(testNotification, [{ value: true }])
    })

    it('unsubscribes connection', async () => {
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
        await assert.rejects(() => client.sendRequest(method, ['']))
    })

    it('unsubscribed connection throws', () => {
        const client = createConnection(createMessagePipe())
        client.listen()
        client.unsubscribe()
        assert.throws(() => client.sendNotification(testNotification))
    })

    it('two listen throw', () => {
        const client = createConnection(createMessagePipe())
        client.listen()
        assert.throws(() => client.listen())
    })

    it('notify on connection unsubscribe', done => {
        const client = createConnection(createMessagePipe())
        client.listen()
        client.onUnsubscribe(() => {
            done()
        })
        client.unsubscribe()
    })

    it('params in notifications', done => {
        const method = 'test'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification(method, params => {
            assert.deepStrictEqual(params, [10, 'vscode'])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification(method, [10, 'vscode'])
    })

    it('params in request/response', async () => {
        const method = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, (params: number[]) => {
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        assert.strictEqual(await client.sendRequest(method, [10, 20, 30]), 60)
    })

    it('params in request/response with signal', async () => {
        const method = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, (params: number[], _signal) => {
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        const signal = new AbortController().signal
        client.listen()
        assert.strictEqual(await client.sendRequest(method, [10, 20, 30], signal), 60)
    })

    it('1 param as array in request', async () => {
        const type = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(type, p1 => {
            assert(Array.isArray(p1))
            assert.strictEqual(p1[0], 10)
            assert.strictEqual(p1[1], 20)
            assert.strictEqual(p1[2], 30)
            return 60
        })
        server.listen()

        const client = createConnection(clientTransports)
        const signal = new AbortController().signal
        client.listen()
        assert.strictEqual(await client.sendRequest(type, [10, 20, 30], signal), 60)
    })

    it('1 param as array in notification', done => {
        const type = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification(type, params => {
            assert.deepStrictEqual(params, [10, 20, 30])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification(type, [10, 20, 30])
    })

    it('untyped request/response', async () => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest('test', (params: number[], _signal) => {
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        const signal = new AbortController().signal
        client.listen()
        assert.strictEqual(await client.sendRequest('test', [10, 20, 30], signal), 60)
    })

    it('untyped notification', done => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification('test', (params: number[], _signal) => {
            assert.deepStrictEqual(params, [10, 20, 30])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification('test', [10, 20, 30])
    })

    it('star request handler', async () => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest((method: string, params: number[], _signal) => {
            assert.strictEqual(method, 'test')
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        const signal = new AbortController().signal
        client.listen()
        assert.strictEqual(await client.sendRequest('test', [10, 20, 30], signal), 60)
    })

    it('star notification handler', done => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification((method: string, params: number[], _signal) => {
            assert.strictEqual(method, 'test')
            assert.deepStrictEqual(params, [10, 20, 30])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification('test', [10, 20, 30])
    })

    it('abort signal is undefined', async () => {
        const type = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(type, (params: number[], _signal) => {
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        assert.strictEqual(await client.sendRequest(type, [10, 20, 30], undefined), 60)
    })

    it('null params in request', async () => {
        const type = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(type, (params, _signal) => 123)
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        ;(client.sendRequest as any)(type, null).then((result: any) => assert.strictEqual(result, 123))
    })

    it('null params in notifications', done => {
        const type = 'test'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification(type, params => {
            assert.strictEqual(params, null)
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        ;(client.sendNotification as any)(type, null)
    })
})
