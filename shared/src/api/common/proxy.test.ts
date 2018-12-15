import assert from 'assert'
import { Observable } from 'rxjs'
import { bufferCount } from 'rxjs/operators'
import { createConnection } from '../protocol/jsonrpc2/connection'
import { createMessageTransports } from '../protocol/jsonrpc2/testHelpers'
import { createProxy, handleRequests } from './proxy'

function createTestProxy<T>(handler: T): Record<keyof T, (...args: any[]) => any> {
    Object.setPrototypeOf(handler, null)
    const [clientTransports, serverTransports] = createMessageTransports()
    const client = createConnection(clientTransports)
    const server = createConnection(serverTransports)

    const proxy = createProxy(client, 'prefix')
    handleRequests(server, 'prefix', handler)
    client.listen()
    server.listen()
    return proxy
}

describe('Proxy', () => {
    describe('proxies calls', () => {
        const proxy = createTestProxy({
            $a: (n: number) => n + 1,
            $b: (n: number) => Promise.resolve(n + 2),
            $c: (m: number, n: number) => m + n,
            $d: (...args: number[]) => args.reduce((sum, n) => sum + n, 0),
            $e: () => Promise.reject('e'),
            $f: () => {
                throw new Error('f')
            },
        })

        test('to functions', async () => expect(await proxy.$a(1)).toBe(2))
        test('to async functions', async () => expect(await proxy.$b(1)).toBe(3))
        test('with multiple arguments ', async () => expect(await proxy.$c(2, 3)).toBe(5))
        test('with variadic arguments ', async () => expect(await proxy.$d(...[2, 3, 4])).toBe(9))
        test('to functions returning a rejected promise', async () => assert.rejects(() => proxy.$e()))
        test('to functions throwing an error', async () => assert.rejects(() => proxy.$f()))
    })

    test('proxies Observables', async () => {
        const proxy = createTestProxy({
            $observe: (...args: number[]) =>
                new Observable<number>(observer => {
                    for (const arg of args) {
                        observer.next(arg + 1)
                    }
                    observer.complete()
                }),
        })

        expect(
            await proxy
                .$observe(1, 2, 3, 4)
                .pipe(bufferCount(4))
                .toPromise()
        ).toEqual([2, 3, 4, 5])
    })
})
