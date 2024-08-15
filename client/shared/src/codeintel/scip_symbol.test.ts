import { describe, expect, test } from 'vitest'

import { parseSymbolName } from './scip_symbol'

describe('parseSymbolName', () => {
    test('basic', () => {
        expect(parseSymbolName('local asdf')).toEqual('test')
    })
})
