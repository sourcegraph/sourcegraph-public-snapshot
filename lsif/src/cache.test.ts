import { GenericCache } from './cache'

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

        const factoryArgs: string[] = []
        const cache = new GenericCache<string, string>(5, () => 1, () => {})

        for (const value of values) {
            const returnValue = await cache.withValue(
                value,
                async () => {
                    factoryArgs.push(value)
                    return value
                },
                async v => v
            )

            expect(returnValue).toBe(value)
        }

        expect(factoryArgs).toEqual(['foo', 'bar', 'baz', 'bonk', 'quux', 'honk', 'baz'])
    })

    it('should asynchronously resolve cache values', async () => {
        const cache = new GenericCache<string, string>(5, () => 1, () => {})

        let innerCalls = 0
        const innerPromise = new Promise<string>(resolve => {
            innerCalls++
            setTimeout(() => resolve('bar'), 10)
        })

        let outerCalls = 0
        const p1 = cache.withValue(
            'foo',
            () => {
                outerCalls++
                return innerPromise
            },
            async v => v
        )
        const p2 = cache.withValue(
            'foo',
            () => {
                outerCalls++
                return innerPromise
            },
            async v => v
        )
        const p3 = cache.withValue(
            'foo',
            () => {
                outerCalls++
                return innerPromise
            },
            async v => v
        )

        expect(await Promise.all([p1, p2, p3])).toEqual(['bar', 'bar', 'bar'])
        expect(innerCalls).toEqual(1)
        expect(outerCalls).toEqual(1)
    })

    it('should call dispose function on eviction', async () => {
        const values = [
            'foo', // foo
            'bar', // bar foo
            'baz', // baz bar (drops foo)
            'foo', // foo baz (drops bar)
        ]

        const disposeArgs: string[] = []
        const cache = new GenericCache<string, string>(
            2,
            () => 1,
            v => {
                disposeArgs.push(v)
            }
        )

        for (const value of values) {
            await cache.withValue(value, async () => value, async v => v)
        }

        // allow disposal to run asynchronously
        await new Promise(resolve =>
            setTimeout(() => {
                expect(disposeArgs).toEqual(['foo', 'bar'])
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

        const factoryArgs: number[] = []
        const cache = new GenericCache<number, number>(5, v => v, () => {})

        for (const value of values) {
            await cache.withValue(
                value,
                async () => {
                    factoryArgs.push(value)
                    return value
                },
                async v => v
            )
        }

        expect(factoryArgs).toEqual([2, 3, 1, 2])
    })

    it('should not evict referenced cache entries', async () => {
        const disposeArgs: string[] = []
        const cache = new GenericCache<string, string>(
            5,
            () => 1,
            v => {
                disposeArgs.push(v)
            }
        )

        await cache.withValue(
            'foo',
            async () => 'foo',
            async () => {
                await cache.withValue(
                    'bar',
                    async () => 'bar',
                    async () => {
                        await cache.withValue(
                            'baz',
                            async () => 'baz',
                            async () => {
                                await cache.withValue(
                                    'bonk',
                                    async () => 'bonk',
                                    async () => {
                                        await cache.withValue(
                                            'quux',
                                            async () => 'quux',
                                            async () => {
                                                // Sixth entry, but nothing to evict (all held)
                                                await cache.withValue(
                                                    'honk',
                                                    async () => 'honk',
                                                    async () => {
                                                        expect(disposeArgs).toEqual([])
                                                    }
                                                )

                                                // Seventh entry, honk can now be removed as it's the least
                                                // recently used value that's not currently under a read lock.
                                                await cache.withValue(
                                                    'ronk',
                                                    async () => 'ronk',
                                                    async () => {
                                                        // allow disposal to run asynchronously
                                                        await new Promise(resolve =>
                                                            setTimeout(() => {
                                                                expect(disposeArgs).toEqual(['honk'])
                                                                resolve()
                                                            }, 10)
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
            }
        )

        // Release and remove the least recently used
        await cache.withValue(
            'honk',
            async () => 'honk',
            async () => {
                // allow disposal to run asynchronously
                await new Promise(resolve =>
                    setTimeout(() => {
                        expect(disposeArgs).toEqual(['honk', 'foo', 'bar'])
                        resolve()
                    }, 10)
                )
            }
        )
    })
})
