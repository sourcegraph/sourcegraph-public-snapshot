import * as assert from 'assert'

import { queryHasCountFilter } from './query-has-count-filter'

describe('queryHasCount()', () => {
    it('returns true when count is specified', () => {
        assert.strictEqual(queryHasCountFilter('const count: 1000'), true)
    })

    it('returns false when count is not specified', () => {
        assert.strictEqual(queryHasCountFilter('const'), false)
    })

    it('returns false when count is escaped', () => {
        assert.strictEqual(queryHasCountFilter('"count:100" lang:ts'), false)
    })

    it('returns false when count is escaped by single quotes', () => {
        assert.strictEqual(queryHasCountFilter("'count:100' lang:ts"), false)
    })
})
