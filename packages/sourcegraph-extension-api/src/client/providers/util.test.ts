import * as assert from 'assert'
import { flattenAndCompact } from './util'

describe('flattenAndCompact', () => {
    it('flattens and compacts 1 level deep', () => {
        assert.deepStrictEqual(flattenAndCompact([null, [1, 2], [3]]), [1, 2, 3])
    })

    it('passes through null', () => {
        assert.deepStrictEqual(flattenAndCompact(null), null)
    })

    it('converts an empty result to null', () => {
        assert.deepStrictEqual(flattenAndCompact([null, [], []]), null)
    })
})
