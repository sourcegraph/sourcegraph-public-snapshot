import assert from 'assert'
import { CancellationTokenSource } from './cancel'
import { createConnection } from './connection'
import { createMessagePipe, createMessageTransports } from './helpers.test'
import { ErrorCodes, ResponseError } from './messages'

describe('Connection', () => {
    it('handle single request', () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, (p1, _token) => {
            assert.strictEqual(p1, 'foo')
            return p1
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        return client.sendRequest(method, 'foo').then(result => assert.strictEqual(result, 'foo'))
    })

    it('handle multiple requests', () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, (p1, _token) => p1)
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        const promises: Promise<string>[] = []
        promises.push(client.sendRequest(method, 'foo'))
        promises.push(client.sendRequest(method, 'bar'))

        return Promise.all(promises).then(values => {
            assert.strictEqual(values.length, 2)
            assert.strictEqual(values[0], 'foo')
            assert.strictEqual(values[1], 'bar')
        })
    })

    it('unhandled request', () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        return client.sendRequest(method, 'foo').then(
            _result => assert.fail('want error'),
            (error: ResponseError<any>) => {
                assert.strictEqual(error.code, ErrorCodes.MethodNotFound)
            }
        )
    })

    it('handler throws an Error', () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, () => {
            throw new Error('test')
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        return client.sendRequest(method, 'foo').then(
            _result => assert.fail('want error'),
            (error: ResponseError<any>) => {
                assert.strictEqual(error.code, ErrorCodes.InternalError)
                assert.strictEqual(error.message, 'test')
                assert.strictEqual(error.data && typeof error.data.stack, 'string')
            }
        )
    })

    it('handler returns a rejected Promise with an Error', () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, () => Promise.reject(new Error('test')))
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        return client.sendRequest(method, 'foo').then(
            _result => assert.fail('want error'),
            (error: ResponseError<any>) => {
                assert.strictEqual(error.code, ErrorCodes.InternalError)
                assert.strictEqual(error.message, 'test')
                assert.strictEqual(error.data && typeof error.data.stack, 'string')
            }
        )
    })

    it('receives undefined request param as null', () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, param => {
            assert.strictEqual(param, null)
            return ''
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        return client.sendRequest(method, undefined)
    })

    it('receives undefined notification param as null', () => {
        const method = 'testNotification'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification(method, param => {
            assert.strictEqual(param, null)
            return ''
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        return client.sendNotification(method, undefined)
    })

    it('receives null as null', () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, param => {
            assert.strictEqual(param, null)
            return null
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        return client.sendRequest(method, null).then(result => {
            assert.strictEqual(result, null)
        })
    })

    it('receives 0 as 0', () => {
        const method = 'test/handleSingleRequest'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, param => {
            assert.strictEqual(param, 0)
            return 0
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        return client.sendRequest(method, 0).then(result => {
            assert.strictEqual(result, 0)
        })
    })

    const testNotification = 'testNotification'
    it('sends and receives notification', done => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification(testNotification, param => {
            assert.strictEqual(param.value, true)
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification(testNotification, { value: true })
    })

    it('unhandled notification event', done => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onUnhandledNotification(message => {
            assert.strictEqual(message.method, testNotification)
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        client.sendNotification(testNotification, { value: true })
    })

    it('unsubscribes connection', () => {
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
        return client.sendRequest(method, '').then(
            _result => assert.fail('want error'),
            () => {
                // noop
            }
        )
    })

    it('unsubscribed connection throws', () => {
        const client = createConnection(createMessagePipe())
        client.listen()
        client.unsubscribe()
        try {
            client.sendNotification(testNotification)
            assert.fail('want error')
        } catch (error) {
            // noop
        }
    })

    it('two listen throw', () => {
        const client = createConnection(createMessagePipe())
        client.listen()
        try {
            client.listen()
            assert.fail('want error')
        } catch (error) {
            // noop
        }
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

    it('params in request/response', () => {
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
        return client.sendRequest(method, [10, 20, 30]).then(result => assert.strictEqual(result, 60))
    })

    it('params in request/response with token', () => {
        const method = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(method, (params: number[], _token) => {
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        const token = new CancellationTokenSource().token
        client.listen()
        return client.sendRequest(method, [10, 20, 30], token).then(result => assert.strictEqual(result, 60))
    })

    it('1 param as array in request', () => {
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
        const token = new CancellationTokenSource().token
        client.listen()
        return client.sendRequest(type, [10, 20, 30], token).then(result => assert.strictEqual(result, 60))
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
        server.onRequest('test', (params: number[], _token) => {
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        const token = new CancellationTokenSource().token
        client.listen()
        const result = await client.sendRequest('test', [10, 20, 30], token)
        assert.strictEqual(result, 60)
    })

    it('untyped notification', done => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification('test', (params: number[], _token) => {
            assert.deepStrictEqual(params, [10, 20, 30])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        const token = new CancellationTokenSource().token
        client.listen()
        client.sendNotification('test', [10, 20, 30], token)
    })

    it('star request handler', () => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest((method: string, params: number[], _token) => {
            assert.strictEqual(method, 'test')
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        const token = new CancellationTokenSource().token
        client.listen()
        return client.sendRequest('test', [10, 20, 30], token).then(result => assert.strictEqual(result, 60))
    })

    it('star notification handler', done => {
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onNotification((method: string, params: number[], _token) => {
            assert.strictEqual(method, 'test')
            assert.deepStrictEqual(params, [10, 20, 30])
            done()
        })
        server.listen()

        const client = createConnection(clientTransports)
        const token = new CancellationTokenSource().token
        client.listen()
        client.sendNotification('test', [10, 20, 30], token)
    })

    it('cancellation token is undefined', () => {
        const type = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(type, (params: number[], _token) => {
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection(clientTransports)
        client.listen()
        return client.sendRequest(type, [10, 20, 30], undefined).then(result => assert.strictEqual(result, 60))
    })

    it('null params in request', () => {
        const type = 'add'
        const [serverTransports, clientTransports] = createMessageTransports()

        const server = createConnection(serverTransports)
        server.onRequest(type, (params, _token) => 123)
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
