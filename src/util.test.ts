import * as assert from 'assert'
import { isEqual } from './util'

describe('isEqual', () => {
    const TESTS: { a: any; b: any; want: boolean }[] = [
        { a: 1, b: 1, want: true },
        { a: 1, b: 2, want: false },
        { a: [1], b: [1], want: true },
        { a: [1], b: [2], want: false },
        { a: { a: 1 }, b: { a: 1 }, want: true },
        { a: { a: 1 }, b: { a: 2 }, want: false },
        { a: { a: [1] }, b: { a: [1] }, want: true },
        { a: { a: [1] }, b: { a: [2] }, want: false },
    ]
    for (const { a, b, want } of TESTS) {
        it(`reports ${JSON.stringify(a)} ${want ? '==' : '!='} ${JSON.stringify(b)}`, () => {
            assert.strictEqual(isEqual(a, b), want)
        })
    }
})
