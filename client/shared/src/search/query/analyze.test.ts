import { describe, expect, it } from 'vitest'

import { SearchPatternType } from '../../graphql-operations'

import { getRelevantTokens } from './analyze'
import { type Node, parseSearchQuery } from './parser'
import { stringHuman } from './printer'
import { scanSearchQuery } from './scanner'
import { Token } from './token'

function parse(input: string, filter = (_node: Node) => true): Token[] {
    const inputPosition = input.indexOf('|')
    if (inputPosition < 0) {
        throw new Error('Query must indicate cursor position via |.')
    }

    input = input.replaceAll('|', '')
    const tokens = scanSearchQuery(input, false, SearchPatternType.standard)
    const result = parseSearchQuery(tokens)
    if (result.type !== 'success') {
        throw new Error(`Expected '${input}' to be a valid query.`)
    }
    return getRelevantTokens(result.node, { start: inputPosition, end: inputPosition }, filter)
}

function annotateToken(token: Token, prefix?: string): string {
    const tokenAnnotation =
        ' '.repeat(token.range.start) +
        '^'.repeat(token.range.end - token.range.start) +
        ` (${prefix ?? ''}${token.type})`

    switch (token.type) {
        case 'filter':
            return [
                annotateToken(token.field, 'filter.field: '),
                token.value ? annotateToken(token.value, 'filter.value: ') : '',
                tokenAnnotation,
            ].join('\n')
        default:
            return tokenAnnotation
    }
}

expect.addSnapshotSerializer({
    serialize: value => stringHuman(value),
    test: () => true,
})

describe('getRelevantTokens', () => {
    describe('branch logic', () => {
        it('preserves sequences', () => {
            expect(parse('a b| c')).toMatchInlineSnapshot(`a b c`)
        })

        it('preserves the target OR branch', () => {
            expect(parse('a b| OR c d')).toMatchInlineSnapshot(`a b`)
            expect(parse('a b OR c| d')).toMatchInlineSnapshot(`c d`)
        })

        it('preserves both AND branches', () => {
            expect(parse('a b| AND c d')).toMatchInlineSnapshot(`(a b AND c d)`)
            expect(parse('a b AND c| d')).toMatchInlineSnapshot(`(a b AND c d)`)
        })

        it('preserves OR branches inside targeted AND branches', () => {
            expect(parse('(a OR b OR c) AND (d| OR e) OR f')).toMatchInlineSnapshot(`((a OR (b OR c)) AND d)`)
        })
    })

    describe('custom filters', () => {
        it('only preserves tokens for which the filter function returns true', () => {
            expect(parse('a b| c', node => node.type === 'pattern' && node.value === 'a')).toMatchInlineSnapshot(`a`)
            expect(
                parse('(a OR b OR c) AND (d| OR e) OR f', node => node.type === 'pattern' && node.value !== 'a')
            ).toMatchInlineSnapshot(`((b OR c) AND d)`)
        })

        it('preserves one branch if the other is filtered out', () => {
            expect(parse('a| AND b', node => node.type === 'pattern' && node.value !== 'a')).toMatchInlineSnapshot(`b`)
        })
    })

    describe('tokens', () => {
        it('preserves quoted filter values', () => {
            expect(parse('a content:"with  space" b| c')).toMatchInlineSnapshot(`a content:"with  space" b c`)
        })

        it('preserves regex literals', () => {
            expect(parse('a /regex/ b| c')).toMatchInlineSnapshot(`a /regex/ b c`)
        })

        it('preserves negated tokens', () => {
            expect(parse('a -repo:exclude/this b| c')).toMatchInlineSnapshot(`a -repo:exclude/this b c`)
        })
    })

    describe('character ranges', () => {
        it('preservers character ranges of patterns and filters and keywords', () => {
            expect.addSnapshotSerializer({
                serialize: tokens =>
                    [
                        stringHuman(tokens),
                        ...(tokens as Token[])
                            .filter(token => ['pattern', 'filter'].includes(token.type))
                            .map(token => annotateToken(token)),
                    ].join('\n'),
                test: () => true,
            })

            expect(parse('abc content:"with  space" def /regex literal/ ghi |')).toMatchInlineSnapshot(`
              abc content:"with  space" def /regex literal/ ghi
              ^^^ (pattern)
                  ^^^^^^^ (filter.field: literal)
                          ^^^^^^^^^^^^^ (filter.value: literal)
                  ^^^^^^^^^^^^^^^^^^^^^ (filter)
                                        ^^^ (pattern)
                                            ^^^^^^^^^^^^^^^ (pattern)
                                                            ^^^ (pattern)
            `)
        })
    })
})
