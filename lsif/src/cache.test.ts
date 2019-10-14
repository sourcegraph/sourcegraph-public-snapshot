import promClient from 'prom-client'
import { Cache, createBarrierPromise } from './cache'
import * as sinon from 'sinon'

describe('Cache', () => {
    const testCacheSizeGauge = new promClient.Gauge({
        name: 'test_cache_size',
        help: 'test_cache_size',
    })

    const testCacheEventsCounter = new promClient.Counter({
        name: 'test_cache_events_total',
        help: 'test_cache_events_total',
        labelNames: ['type'],
    })

    const testMetrics = {
        sizeGauge: testCacheSizeGauge,
        eventsCounter: testCacheEventsCounter,
    }

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

        const factory = sinon.stub<string[], Promise<string>>()
        for (const [i, value] of expectedInstantiations.entries()) {
            // Log the value arg and resolve the cache data immediately
            factory.onCall(i).returns(Promise.resolve(value))
        }

        const cache = new Cache<string, string>(5, () => 1, () => {}, 0, testMetrics)
        for (const value of values) {
            const returnValue = await cache.withValue(value, () => factory(value), () => true, v => Promise.resolve(v))
            expect(returnValue).toBe(value)
        }

        // Expect the args of the factory to equal the resolved values
        expect(factory.args).toEqual(expectedInstantiations.map(v => [v]))
    })

    it('should asynchronously resolve cache values', async () => {
        const factory = sinon.stub<string[], Promise<string>>()
        const { wait, done } = createBarrierPromise()
        factory.returns(wait.then(() => 'bar'))

        const cache = new Cache<string, string>(5, () => 1, () => {}, 0, testMetrics)
        const p1 = cache.withValue('foo', factory, () => true, v => Promise.resolve(v))
        const p2 = cache.withValue('foo', factory, () => true, v => Promise.resolve(v))
        const p3 = cache.withValue('foo', factory, () => true, v => Promise.resolve(v))
        done()

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

        const { wait, done } = createBarrierPromise()
        const disposer = sinon.spy(done)
        const cache = new Cache<string, string>(2, () => 1, disposer, 0, testMetrics)

        for (const value of values) {
            await cache.withValue(value, () => Promise.resolve(value), () => true, v => Promise.resolve(v))
        }

        await wait
        expect(disposer.args).toEqual([['foo'], ['bar']])
    })

    it('should calculate size by resolved value', async () => {
        const values = [
            2, // 2,   size = 2
            3, // 3 2, size = 5
            1, // 1 3, size = 4
            2, // 1 2, size = 3
        ]

        const expectedInstantiations = [2, 3, 1, 2]

        const factory = sinon.stub<number[], Promise<number>>()
        for (const [i, value] of expectedInstantiations.entries()) {
            factory.onCall(i).returns(Promise.resolve(value))
        }

        const cache = new Cache<number, number>(5, v => v, () => {}, 0, testMetrics)
        for (const value of values) {
            await cache.withValue(value, () => factory(value), () => true, v => Promise.resolve(v))
        }

        expect(factory.args).toEqual(expectedInstantiations.map(v => [v]))
    })

    it('should not evict referenced cache entries', async () => {
        const { wait, done } = createBarrierPromise()
        const disposer = sinon.spy(done)
        const cache = new Cache<string, string>(5, () => 1, disposer, 0, testMetrics)

        const fooFactory = () => Promise.resolve('foo')
        const barFactory = () => Promise.resolve('bar')
        const bazFactory = () => Promise.resolve('baz')
        const bonkFactory = () => Promise.resolve('bonk')
        const quuxFactory = () => Promise.resolve('quux')
        const honkFactory = () => Promise.resolve('honk')
        const ronkFactory = () => Promise.resolve('ronk')
        const valid = () => true

        await cache.withValue('foo', fooFactory, valid, async () => {
            await cache.withValue('bar', barFactory, valid, async () => {
                await cache.withValue('baz', bazFactory, valid, async () => {
                    await cache.withValue('bonk', bonkFactory, valid, async () => {
                        await cache.withValue('quux', quuxFactory, valid, async () => {
                            // Sixth entry, but nothing to evict (all held)
                            await cache.withValue('honk', honkFactory, valid, () => Promise.resolve())

                            // Seventh entry, honk can now be removed as it's the least
                            // recently used value that's not currently under a read lock.
                            await cache.withValue('ronk', ronkFactory, valid, () => Promise.resolve())
                        })
                    })
                })
            })
        })

        // Release and remove the least recently used

        await cache.withValue(
            'honk',
            () => Promise.resolve('honk'),
            () => true,
            async () => {
                await wait
                expect(disposer.args).toEqual([['honk'], ['foo'], ['bar']])
            }
        )
    })

    it('should dispose busted keys', async () => {
        const { wait, done } = createBarrierPromise()
        const disposer = sinon.spy(done)
        const cache = new Cache<string, string>(5, () => 1, disposer, 0, testMetrics)

        const factory = sinon.stub<string[], Promise<string>>()
        factory.returns(Promise.resolve('foo'))

        // Construct then bust a same key
        await cache.withValue('foo', factory, () => true, () => Promise.resolve())
        await cache.bustKey('foo')
        await wait

        // Ensure value was disposed
        expect(disposer.args).toEqual([['foo']])

        // Ensure entry was removed
        expect(cache.withValue('foo', factory, () => true, () => Promise.resolve()))
        expect(factory.args).toHaveLength(2)
    })

    it('should wait to dispose busted keys that are in use', async () => {
        const { wait: wait1, done: done1 } = createBarrierPromise()
        const { wait: wait2, done: done2 } = createBarrierPromise()

        const factory = () => Promise.resolve('foo')
        const disposer = sinon.spy(done1)
        const cache = new Cache<string, string>(5, () => 1, disposer, 0, testMetrics)

        // Create a cache entry for 'foo' that blocks on done2
        const p1 = cache.withValue('foo', factory, () => true, () => wait2)

        // Attempt to bust the cache key that's used in the blocking promise above
        const p2 = cache.bustKey('foo')

        // Ensure that p1 and p2 are blocked on each other
        const timedFactory = new Promise(resolve => setTimeout(() => resolve('$'), 10))
        const winner = await Promise.race([p1, p2, timedFactory])
        expect(winner).toEqual('$')

        // Ensure dispose hasn't been called
        expect(disposer.args).toHaveLength(0)

        // Unblock p1
        done2()

        // Show that all promises are unblocked and dispose was called
        await Promise.all([p1, p2, wait1])
        expect(disposer.args).toEqual([['foo']])
    })

    it('should revalidate entries', async () => {
        const factory = sinon.stub<[], Promise<string>>()
        factory.returns(Promise.resolve('bar'))
        const disposer = sinon.spy()
        const validator = sinon.stub<[number], boolean>()
        validator.returns(true)

        let currentTime = 1000
        const cache = new Cache<string, string>(5, () => 1, disposer, 100, testMetrics, () => currentTime)

        // First construction
        await cache.withValue('foo', factory, validator, () => Promise.resolve())
        expect(factory.callCount).toEqual(1)
        expect(validator.args).toEqual([])

        // Within validity interval
        currentTime += 50 // eslint-disable-line require-atomic-updates
        await cache.withValue('foo', factory, validator, () => Promise.resolve())
        expect(factory.callCount).toEqual(1)
        expect(validator.args).toEqual([])

        // Outside of validity interval
        currentTime += 82 // eslint-disable-line require-atomic-updates
        await cache.withValue('foo', factory, validator, () => Promise.resolve())
        expect(factory.callCount).toEqual(1)
        expect(validator.args).toEqual([[1000]])

        // Within next validity interval
        currentTime += 10 // eslint-disable-line require-atomic-updates
        await cache.withValue('foo', factory, validator, () => Promise.resolve())
        expect(factory.callCount).toEqual(1)
        expect(validator.args).toEqual([[1000]])

        // Outside of next validity interval
        currentTime += 91 // eslint-disable-line require-atomic-updates
        await cache.withValue('foo', factory, validator, () => Promise.resolve())
        expect(factory.callCount).toEqual(1)
        expect(validator.args).toEqual([[1000], [1132]])
    })

    it('should recreate invalid entries', async () => {
        const factory = sinon.stub<[], Promise<string>>()
        factory.returns(Promise.resolve('bar'))
        const disposer = sinon.spy()
        const validator = sinon.stub<[number], boolean>()
        validator.returns(false)

        let currentTime = 1000
        const cache = new Cache<string, string>(5, () => 1, disposer, 100, testMetrics, () => currentTime)

        // First construction
        await cache.withValue('foo', factory, validator, () => Promise.resolve())
        expect(factory.callCount).toEqual(1)
        expect(validator.args).toEqual([])

        // Within validity interval
        currentTime += 50 // eslint-disable-line require-atomic-updates
        await cache.withValue('foo', factory, validator, () => Promise.resolve())
        expect(factory.callCount).toEqual(1)
        expect(validator.args).toEqual([])

        // Outside of validity interval
        currentTime += 82 // eslint-disable-line require-atomic-updates
        await cache.withValue('foo', factory, validator, () => Promise.resolve())
        expect(factory.callCount).toEqual(2)
        expect(validator.args).toEqual([[1000]])
    })
})
