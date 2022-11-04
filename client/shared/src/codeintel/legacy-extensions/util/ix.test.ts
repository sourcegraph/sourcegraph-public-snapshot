import * as assert from 'assert'

import { asyncGeneratorFromPromise, concat, observableFromAsyncIterator } from './ix'

describe('observableFromAsyncIterator', () => {
    it('converts iterator into an observable', async () => {
        const observable = observableFromAsyncIterator(() =>
            (async function* (): AsyncIterator<number> {
                await Promise.resolve()
                yield 1
                yield 2
                yield 3
                yield 4
                yield 5
            })()
        )

        const values: number[] = []
        await new Promise<void>(complete => observable.subscribe({ next: value => values.push(value), complete }))
        assert.deepStrictEqual(values, [1, 2, 3, 4, 5])
    })

    it('throws iterator error', async () => {
        const observable = observableFromAsyncIterator(() =>
            (async function* (): AsyncIterator<number> {
                await Promise.resolve()
                yield 1
                yield 2
                yield 3
                throw new Error('oops')
            })()
        )

        const error = await new Promise(error => observable.subscribe({ error }))
        assert.deepStrictEqual(error, new Error('oops'))
    })
})

describe('concat', () => {
    it('returns all previous values', async () => {
        const iterable = concat(
            (async function* (): AsyncIterable<number[] | null> {
                await Promise.resolve()
                yield [1]
                yield [2, 3]
                yield [4, 5]
            })()
        )

        assert.deepStrictEqual(await gatherValues(iterable), [[1], [1, 2, 3], [1, 2, 3, 4, 5]])
    })

    it('ignores nulls', async () => {
        const iterable = concat(
            (async function* (): AsyncIterable<number[] | null> {
                await Promise.resolve()
                yield null
                yield [1]
                yield null
                yield [2, 3]
                yield [4, 5]
                yield null
            })()
        )

        assert.deepStrictEqual(await gatherValues(iterable), [[1], [1, 2, 3], [1, 2, 3, 4, 5]])
    })
})

describe('asyncGeneratorFromPromise', () => {
    it('yields mapped values', async () => {
        const iterable = asyncGeneratorFromPromise(async (value: number) => Promise.resolve(value * 2))

        assert.deepStrictEqual(await gatherValues(iterable(24)), [48])
    })
})

async function gatherValues<T>(iterable: AsyncIterable<T>): Promise<T[]> {
    const values = []
    for await (const value of iterable) {
        values.push(value)
    }

    return values
}
