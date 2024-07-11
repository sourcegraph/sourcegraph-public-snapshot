import { expect, describe, it } from 'vitest'

import { escapeQuery_TEST_ONLY as escapeQuery } from './sources'

describe('escapeQuery', () => {
    it.each([
        ['repo:sourcegraph', '"repo:sourcegraph"'],
        ['file:main.go', '"file:main.go"'],
        ['r:sourcegraph f:main.go', '"r:sourcegraph" "f:main.go"'],
        ['OR AND NOT', '"OR" "AND" "NOT"'],
        ['( foo )', '"(" "foo" ")"'],
        ['(foo OR bar) AND baz', '"(foo" "OR" "bar)" "AND" "baz"'],
    ])('escapes special tokens: %s -> %s', (query, expected) => {
        expect(escapeQuery(query)).toBe(expected)
    })

    it('preserves regex patterns', () => {
        expect(escapeQuery('repo:^sourcegraph /f.o$/ bar')).toBe('"repo:^sourcegraph" /f.o$/ "bar"')
    })

    it('escapes quotes in patterns', () => {
        expect(escapeQuery('foo"bar')).toBe('"foo\\"bar"')
    })
})
