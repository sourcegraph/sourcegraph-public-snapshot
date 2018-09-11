import assert from 'assert'
import { Duplex } from 'stream'
import { inherits } from 'util'
import { CancellationTokenSource } from './cancel'
import { createConnection } from './connection'
import { ErrorCodes, NotificationType, RequestType, ResponseError } from './messages'
import { StreamMessageReader, StreamMessageWriter } from './transports/stream'

interface TestDuplex extends Duplex {}

interface TestDuplexConstructor {
    new (name?: string, dbg?: boolean): TestDuplex
}

const TestDuplex: TestDuplexConstructor = ((): TestDuplexConstructor => {
    function TestDuplex(this: any, name = 'ds1', dbg = false): any {
        Duplex.call(this)
        this.name = name
        this.dbg = dbg
    }
    inherits(TestDuplex, Duplex)
    TestDuplex.prototype._write = function(
        this: any,
        chunk: string | Buffer,
        _encoding: string,
        done: () => void
    ): void {
        if (this.dbg) {
            console.log(this.name + ': write: ' + chunk.toString())
        }
        setImmediate(() => {
            this.emit('data', chunk)
        })
        done()
    }
    TestDuplex.prototype._read = function(this: any, _size: number): void {
        /* noop */
    }
    return (TestDuplex as any) as TestDuplexConstructor
})()

// tslint:disable no-floating-promises
describe('Connection', () => {
    it('duplex stream', done => {
        const stream = new TestDuplex('ds1')
        stream.on('data', chunk => {
            assert.strictEqual('Hello World', chunk.toString())
            done()
        })
        stream.write('Hello World')
    })

    it('duplex stream connection', done => {
        const type = new RequestType<string, string, void, void>('test/handleSingleRequest')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const connection = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        connection.listen()
        let counter = 0
        let content = ''
        duplexStream2.on('data', chunk => {
            content += chunk.toString()
            if (++counter === 2) {
                assert.strictEqual(content.indexOf('Content-Length: 75'), 0)
                done()
            }
        })
        connection.sendRequest(type, 'foo')
    })

    it('handle single request', () => {
        const type = new RequestType<string, string, void, void>('test/handleSingleRequest')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onRequest(type, (p1, _token) => {
            assert.strictEqual(p1, 'foo')
            return p1
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        return client.sendRequest(type, 'foo').then(result => assert.strictEqual(result, 'foo'))
    })

    it('handle multiple requests', () => {
        const type = new RequestType<string, string, void, void>('test/handleSingleRequest')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onRequest(type, (p1, _token) => p1)
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        const promises: Promise<string>[] = []
        promises.push(client.sendRequest(type, 'foo'))
        promises.push(client.sendRequest(type, 'bar'))

        return Promise.all(promises).then(values => {
            assert.strictEqual(values.length, 2)
            assert.strictEqual(values[0], 'foo')
            assert.strictEqual(values[1], 'bar')
        })
    })

    it('unhandled request', () => {
        const type = new RequestType<string, string, void, void>('test/handleSingleRequest')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        return client.sendRequest(type, 'foo').then(
            _result => assert.fail('want error'),
            (error: ResponseError<any>) => {
                assert.strictEqual(error.code, ErrorCodes.MethodNotFound)
            }
        )
    })

    it('receives undefined request param as null', () => {
        const type = new RequestType<string, string, void, void>('test/handleSingleRequest')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onRequest(type, param => {
            assert.strictEqual(param, null)
            return ''
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        return client.sendRequest(type, undefined)
    })

    it('receives undefined notification param as null', () => {
        const type = new NotificationType<string, void>('testNotification')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onNotification(type, param => {
            assert.strictEqual(param, null)
            return ''
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        return client.sendNotification(type, undefined)
    })

    it('receives null as null', () => {
        const type = new RequestType<string | null, string | null, void, void>('test/handleSingleRequest')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onRequest(type, param => {
            assert.strictEqual(param, null)
            return null
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        return client.sendRequest(type, null).then(result => {
            assert.strictEqual(result, null)
        })
    })

    it('receives 0 as 0', () => {
        const type = new RequestType<number, number, void, void>('test/handleSingleRequest')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onRequest(type, param => {
            assert.strictEqual(param, 0)
            return 0
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        return client.sendRequest(type, 0).then(result => {
            assert.strictEqual(result, 0)
        })
    })

    const testNotification = new NotificationType<{ value: boolean }, void>('testNotification')
    it('sends and receives notification', done => {
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onNotification(testNotification, param => {
            assert.strictEqual(param.value, true)
            done()
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        client.sendNotification(testNotification, { value: true })
    })

    it('unhandled notification event', done => {
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onUnhandledNotification(message => {
            assert.strictEqual(message.method, testNotification.method)
            done()
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        client.sendNotification(testNotification, { value: true })
    })

    it('unsubscribes connection', () => {
        const type = new RequestType<string | null, string | null, void, void>('test/handleSingleRequest')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onRequest(type, _param => {
            client.unsubscribe()
            return ''
        })
        server.listen()

        client.listen()
        return client.sendRequest(type, '').then(
            _result => assert.fail('want error'),
            () => {
                /* noop */
            }
        )
    })

    it('unsubscribed connection throws', () => {
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        client.unsubscribe()
        try {
            client.sendNotification(testNotification)
            assert.fail('want error')
        } catch (error) {
            /* noop */
        }
    })

    it('two listen throw', () => {
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        try {
            client.listen()
            assert.fail('want error')
        } catch (error) {
            /* noop */
        }
    })

    it('notify on connection unsubscribe', done => {
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        client.onUnsubscribe(() => {
            done()
        })
        client.unsubscribe()
    })

    it('params in notifications', done => {
        const type = new NotificationType<[number, string], void>('test')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onNotification(type, params => {
            assert.deepStrictEqual(params, [10, 'vscode'])
            done()
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        client.sendNotification(type, [10, 'vscode'])
    })

    it('params in request/response', () => {
        const type = new RequestType<[number, number, number], number, void, void>('add')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onRequest(type, params => {
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        return client.sendRequest(type, [10, 20, 30]).then(result => assert.strictEqual(result, 60))
    })

    it('params in request/response with token', () => {
        const type = new RequestType<[number, number, number], number, void, void>('add')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onRequest(type, (params, _token) => {
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        const token = new CancellationTokenSource().token
        client.listen()
        return client.sendRequest(type, [10, 20, 30], token).then(result => assert.strictEqual(result, 60))
    })

    it('1 param as array in request', () => {
        const type = new RequestType<number[], number, void, void>('add')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onRequest(type, p1 => {
            assert(Array.isArray(p1))
            assert.strictEqual(p1[0], 10)
            assert.strictEqual(p1[1], 20)
            assert.strictEqual(p1[2], 30)
            return 60
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        const token = new CancellationTokenSource().token
        client.listen()
        return client.sendRequest(type, [10, 20, 30], token).then(result => assert.strictEqual(result, 60))
    })

    it('1 param as array in notification', done => {
        const type = new NotificationType<number[], void>('add')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onNotification(type, params => {
            assert.deepStrictEqual(params, [10, 20, 30])
            done()
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        client.sendNotification(type, [10, 20, 30])
    })

    it('untyped request/response', () => {
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onRequest('test', (params: number[], _token) => {
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        const token = new CancellationTokenSource().token
        client.listen()
        client.sendRequest('test', [10, 20, 30], token).then(result => assert.strictEqual(result, 60))
    })

    it('untyped notification', done => {
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onNotification('test', (params: number[], _token) => {
            assert.deepStrictEqual(params, [10, 20, 30])
            done()
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        const token = new CancellationTokenSource().token
        client.listen()
        client.sendNotification('test', [10, 20, 30], token)
    })

    it('star request handler', () => {
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onRequest((method: string, params: number[], _token) => {
            assert.strictEqual(method, 'test')
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        const token = new CancellationTokenSource().token
        client.listen()
        return client.sendRequest('test', [10, 20, 30], token).then(result => assert.strictEqual(result, 60))
    })

    it('star notification handler', done => {
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onNotification((method: string, params: number[], _token) => {
            assert.strictEqual(method, 'test')
            assert.deepStrictEqual(params, [10, 20, 30])
            done()
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        const token = new CancellationTokenSource().token
        client.listen()
        client.sendNotification('test', [10, 20, 30], token)
    })

    it('cancellation token is undefined', () => {
        const type = new RequestType<[number, number, number], number, void, void>('add')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onRequest(type, (params, _token) => {
            assert.deepStrictEqual(params, [10, 20, 30])
            return params.reduce((sum, n) => sum + n, 0)
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        return client.sendRequest(type, [10, 20, 30], undefined).then(result => assert.strictEqual(result, 60))
    })

    it('null params in request', () => {
        const type = new RequestType<string | null, number, void, void>('add')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onRequest(type, (params, _token) => 123)
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        ;(client.sendRequest as any)(type, null).then((result: any) => assert.strictEqual(result, 123))
    })

    it('null params in notifications', done => {
        const type = new NotificationType<string | null, void>('test')
        const duplexStream1 = new TestDuplex('ds1')
        const duplexStream2 = new TestDuplex('ds2')

        const server = createConnection({
            reader: new StreamMessageReader(duplexStream2),
            writer: new StreamMessageWriter(duplexStream1),
        })
        server.onNotification(type, params => {
            assert.strictEqual(params, null)
            done()
        })
        server.listen()

        const client = createConnection({
            reader: new StreamMessageReader(duplexStream1),
            writer: new StreamMessageWriter(duplexStream2),
        })
        client.listen()
        ;(client.sendNotification as any)(type, null)
    })
})
