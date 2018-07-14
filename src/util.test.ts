import * as assert from 'assert'
import { compact, flatten, isEqual, isFunction } from './util'

describe('isFunction', () => {
    it('reports true for functions', () => {
        assert.ok(isFunction(() => void 0))
        assert.ok(
            // tslint:disable-next-line:only-arrow-functions
            isFunction(function(): void {
                /* noop */
            })
        )
        assert.ok(isFunction({ f: () => void 0 }.f))
    })

    it('reports true for non-functions', () => {
        assert.strictEqual(isFunction('[object Function]'), false)
        assert.strictEqual(isFunction({}), false)
    })
})

describe('flatten', () => {
    it('flattens arrays 1 level deep', () => {
        assert.deepStrictEqual(flatten([]), [])
        assert.deepStrictEqual(flatten([1]), [1])
        assert.deepStrictEqual(flatten([1, 'a']), [1, 'a'])
        assert.deepStrictEqual(flatten([[1], 2, [3]]), [1, 2, 3])
        assert.deepStrictEqual(flatten([[1, 2], [[3], 4], [5, [6]]]), [1, 2, [3], 4, 5, [6]])
    })
})

describe('compact', () => {
    it('removes falsey values from arrays', () => {
        assert.deepStrictEqual(compact([]), [])
        assert.deepStrictEqual(compact([0, 1, '', true, false, NaN]), [1, true])
    })
})

describe('isEqual', () => {
    const TESTS: { a: any; b: any; want: boolean }[] = [
        { a: 1, b: 1, want: true },
        { a: 1, b: 2, want: false },
        { a: {}, b: undefined, want: false },
        { a: undefined, b: {}, want: false },
        { a: [1], b: [1], want: true },
        { a: [1], b: [2], want: false },
        { a: [1], b: [1, 1], want: false },
        { a: { a: 1 }, b: { a: 1 }, want: true },
        { a: { a: 1 }, b: { b: 1 }, want: false },
        { a: { a: 1 }, b: { a: 2 }, want: false },
        { a: { a: 1 }, b: { a: 1, b: 2 }, want: false },
        { a: { a: [1] }, b: { a: [1] }, want: true },
        { a: { a: [1] }, b: { a: [2] }, want: false },
    ]
    for (const { a, b, want } of TESTS) {
        it(`reports ${JSON.stringify(a)} ${want ? '==' : '!='} ${JSON.stringify(b)}`, () => {
            assert.strictEqual(isEqual(a, b), want)
        })
    }
})
