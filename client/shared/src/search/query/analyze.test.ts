import { describe, expect, it } from 'vitest'

import { getRelevantTokens } from './analyze'
import { type Node, parseSearchQuery } from './parser'
import { stringHuman } from './printer'

export const parse = (input: string, filter = (_node: Node) => true): string => {
    const inputPosition = input.indexOf('|')
    if (inputPosition < 0) {
        throw new Error('Query must indicate cursor position via |.')
    }

    input = input.replaceAll('|', '')
    const result = parseSearchQuery(input)
    if (result.type !== 'success') {
        throw new Error(`Expected '${input}' to be a valid query.`)
    }
    return stringHuman(getRelevantTokens(result.node, { start: inputPosition, end: inputPosition }, filter))
}

describe('getRelevantTokens', () => {
    describe('branch logic', () => {
        it('preserves sequences', () => {
            expect(parse('a b| c')).toMatchInlineSnapshot('"a b c"')
        })

        it('preserves the target OR branch', () => {
            expect(parse('a b| OR c d')).toMatchInlineSnapshot('"a b"')
            expect(parse('a b OR c| d')).toMatchInlineSnapshot('"c d"')
        })

        it('preserves both AND branches', () => {
            expect(parse('a b| AND c d')).toMatchInlineSnapshot('"(a b AND c d)"')
            expect(parse('a b AND c| d')).toMatchInlineSnapshot('"(a b AND c d)"')
        })

        it('preserves OR branches inside targeted AND branches', () => {
            expect(parse('(a OR b OR c) AND (d| OR e) OR f')).toMatchInlineSnapshot('"((a OR (b OR c)) AND d)"')
        })
    })

    describe('custom filters', () => {
        it('only preserves tokens for which the filter function returns true', () => {
            expect(parse('a b| c', node => node.type === 'pattern' && node.value === 'a')).toMatchInlineSnapshot('"a"')
            expect(
                parse('(a OR b OR c) AND (d| OR e) OR f', node => node.type === 'pattern' && node.value !== 'a')
            ).toMatchInlineSnapshot('"((b OR c) AND d)"')
        })

        it('preserves one branch if the other is filtered out', () => {
            expect(parse('a| AND b', node => node.type === 'pattern' && node.value !== 'a')).toMatchInlineSnapshot(
                '"b"'
            )
        })
    })

    describe('tokens', () => {
        it('preserves quoted filter values', () => {
            expect(parse('a content:"with  space" b| c')).toMatchInlineSnapshot('"a content:\\"with  space\\" b c"')
        })

        it('preserves regex literals', () => {
            expect(parse('a /regex/ b| c')).toMatchInlineSnapshot('"a /regex/ b c"')
        })

        it('preserves negated tokens', () => {
            expect(parse('a -repo:exclude/this b| c')).toMatchInlineSnapshot('"a -repo:exclude/this b c"')
        })
    })
})
