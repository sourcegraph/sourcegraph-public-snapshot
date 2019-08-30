import { GenericCache } from './cache'
import * as sinon from 'sinon'

describe('GenericCache', () => {
    it('should evict items based by reverse recency', async () => {
        const values = [
            'foo', // foo*
            'bar', // bar* foo
            'baz', // baz* bar foo
            'bonk', // bonk* baz bar foo
            'quux', // quux* bonk baz bar foo
            'bar', // bar quux bonk baz foo
            'foo', // foo bar quux bonk baz
            'honk', // honk* foo bar quux bonk
            'foo', // foo honk bar quux bonk
            'baz', // baz* foo honk bar quux
        ]

        // These are the cache values that need to be created, in-order
        const expectedInstantiations = ['foo', 'bar', 'baz', 'bonk', 'quux', 'honk', 'baz']

        let i = 0
        const factory = sinon.stub()
        for (const value of expectedInstantiations) {
            // Log the value arg and resolve the cache data immediately
            factory.onCall(i++).returns(Promise.resolve(value))
        }

        const cache = new GenericCache<string, string>(5, () => 1, () => {})
        for (const value of values) {
            const returnValue = await cache.withValue(value, () => factory(value), v => Promise.resolve(v))
            expect(returnValue).toBe(value)
        }

        // Expect the args of the factory to equal the resolved values
        expect(factory.args).toEqual(expectedInstantiations.map(v => [v]))
    })

    it('should asynchronously resolve cache values', async () => {
        const factory = sinon.stub()
        factory.returns(
            new Promise<string>(resolve => {
                setTimeout(() => resolve('bar'), 10)
            })
        )

        const cache = new GenericCache<string, string>(5, () => 1, () => {})
        const p1 = cache.withValue('foo', factory, v => Promise.resolve(v))
        const p2 = cache.withValue('foo', factory, v => Promise.resolve(v))
        const p3 = cache.withValue('foo', factory, v => Promise.resolve(v))

        expect(await Promise.all([p1, p2, p3])).toEqual(['bar', 'bar', 'bar'])
        expect(factory.callCount).toEqual(1)
    })

    it('should call dispose function on eviction', async () => {
        const values = [
            'foo', // foo
            'bar', // bar foo
            'baz', // baz bar (drops foo)
            'foo', // foo baz (drops bar)
        ]

        const disposer = sinon.spy()
        const cache = new GenericCache<string, string>(2, () => 1, disposer)

        for (const value of values) {
            await cache.withValue(value, () => Promise.resolve(value), v => Promise.resolve(v))
        }

        // allow disposal to run asynchronously
        await new Promise(resolve =>
            setTimeout(() => {
                expect(disposer.args).toEqual([['foo'], ['bar']])
                resolve()
            }, 10)
        )
    })

    it('should calculate size by resolved value', async () => {
        const values = [
            2, // 2,   size = 2
            3, // 3 2, size = 5
            1, // 1 3, size = 4
            2, // 1 2, size = 3
        ]

        const expectedInstantiations = [2, 3, 1, 2]

        let i = 0
        const factory = sinon.stub()
        for (const value of expectedInstantiations) {
            factory.onCall(i++).returns(Promise.resolve(value))
        }

        const cache = new GenericCache<number, number>(5, v => v, () => {})
        for (const value of values) {
            await cache.withValue(value, () => factory(value), v => Promise.resolve(v))
        }

        expect(factory.args).toEqual(expectedInstantiations.map(v => [v]))
    })

    it('should not evict referenced cache entries', async () => {
        const disposer = sinon.spy()
        const cache = new GenericCache<string, string>(5, () => 1, disposer)

        const assertDisposeCalls = async (...expected: string[]) => {
            await new Promise(resolve =>
                setTimeout(() => {
                    expect(disposer.args).toEqual(expected.map(v => [v]))
                    resolve()
                }, 10)
            )
        }

        await cache.withValue(
            'foo',
            () => Promise.resolve('foo'),
            async () => {
                await cache.withValue(
                    'bar',
                    () => Promise.resolve('bar'),
                    async () => {
                        await cache.withValue(
                            'baz',
                            () => Promise.resolve('baz'),
                            async () => {
                                await cache.withValue(
                                    'bonk',
                                    () => Promise.resolve('bonk'),
                                    async () => {
                                        await cache.withValue(
                                            'quux',
                                            () => Promise.resolve('quux'),
                                            async () => {
                                                // Sixth entry, but nothing to evict (all held)
                                                await cache.withValue(
                                                    'honk',
                                                    () => Promise.resolve('honk'),
                                                    () => assertDisposeCalls()
                                                )

                                                // Seventh entry, honk can now be removed as it's the least
                                                // recently used value that's not currently under a read lock.
                                                await cache.withValue(
                                                    'ronk',
                                                    () => Promise.resolve('ronk'),
                                                    () => assertDisposeCalls('honk')
                                                )
                                            }
                                        )
                                    }
                                )
                            }
                        )
                    }
                )
            }
        )

        // Release and remove the least recently used
        await cache.withValue('honk', () => Promise.resolve('honk'), () => assertDisposeCalls('honk', 'foo', 'bar'))
    })
})
