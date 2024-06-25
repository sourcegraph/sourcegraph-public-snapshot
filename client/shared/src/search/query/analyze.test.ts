import { describe, expect, it } from 'vitest'

import { SearchPatternType } from '../../graphql-operations'

import { getRelevantTokens } from './analyze'
import { type Node, parseSearchQuery } from './parser'
import { stringHuman } from './printer'
import { scanSearchQuery } from './scanner'
import { Token } from './token'

function parse(input: string, filter = (_node: Node) => true): ReturnType<typeof getRelevantTokens> {
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

function annotateToken(token: Token, prefix?: string, sourceMap?: Map<Token, Token['range']>): string {
    const range = sourceMap?.get(token) ?? token.range
    const tokenAnnotation =
        ' '.repeat(range.start) + '^'.repeat(range.end - range.start) + ` (${prefix ?? ''}${token.type})`

    switch (token.type) {
        case 'filter': {
            return [
                annotateToken(token.field, 'filter.field: ', sourceMap),
                token.value ? annotateToken(token.value, 'filter.value: ', sourceMap) : '',
                tokenAnnotation,
            ].join('\n')
        }
        default: {
            return tokenAnnotation
        }
    }
}

expect.addSnapshotSerializer({
    serialize: value => stringHuman(value.tokens),
    test: () => true,
})

describe('getRelevantTokens', () => {
    describe('branch logic', () => {
        it('preserves sequences', () => {
            expect(parse('a b| c')).toMatchInlineSnapshot('a b c')
        })

        it('preserves the target OR branch', () => {
            expect(parse('a b| OR c d')).toMatchInlineSnapshot('a b')
            expect(parse('a b OR c| d')).toMatchInlineSnapshot('c d')
        })

        it('preserves both AND branches', () => {
            expect(parse('a b| AND c d')).toMatchInlineSnapshot('(a b AND c d)')
            expect(parse('a b AND c| d')).toMatchInlineSnapshot('(a b AND c d)')
        })

        it('preserves OR branches inside targeted AND branches', () => {
            expect(parse('(a OR b OR c) AND (d| OR e) OR f')).toMatchInlineSnapshot('((a OR (b OR c)) AND d)')
        })
    })

    describe('custom filters', () => {
        it('only preserves tokens for which the filter function returns true', () => {
            expect(parse('a b| c', node => node.type === 'pattern' && node.value === 'a')).toMatchInlineSnapshot('a')
            expect(
                parse('(a OR b OR c) AND (d| OR e) OR f', node => node.type === 'pattern' && node.value !== 'a')
            ).toMatchInlineSnapshot('((b OR c) AND d)')
        })

        it('preserves one branch if the other is filtered out', () => {
            expect(parse('a| AND b', node => node.type === 'pattern' && node.value !== 'a')).toMatchInlineSnapshot('b')
        })
    })

    describe('tokens', () => {
        it('preserves quoted filter values', () => {
            expect(parse('a content:"with  space" b| c')).toMatchInlineSnapshot('a content:"with  space" b c')
        })

        it('preserves regex literals', () => {
            expect(parse('a /regex/ b| c')).toMatchInlineSnapshot('a /regex/ b c')
        })

        it('preserves negated tokens', () => {
            expect(parse('a -repo:exclude/this b| c')).toMatchInlineSnapshot('a -repo:exclude/this b c')
        })
    })

    describe('character ranges', () => {
        it('generates proper character ranges for returned tokens', () => {
            expect.addSnapshotSerializer({
                serialize: result =>
                    [stringHuman(result.tokens), ...(result.tokens as Token[]).map(token => annotateToken(token))].join(
                        '\n'
                    ),
                test: () => true,
            })

            expect(parse('abc AND content:"with  space" def| OR /regex literal/ ghi')).toMatchInlineSnapshot(`
              (abc AND content:"with  space" def)
              ^ (openingParen)
               ^^^ (pattern)
                  ^ (whitespace)
                   ^^^ (keyword)
                      ^ (whitespace)
                       ^^^^^^^ (filter.field: literal)
                               ^^^^^^^^^^^^^ (filter.value: literal)
                       ^^^^^^^^^^^^^^^^^^^^^ (filter)
                                            ^ (whitespace)
                                             ^^^ (pattern)
                                                ^ (closingParen)
            `)
        })
    })

    it('maps tokens to their original positions in the query', () => {
        const input = 'abc AND content:"with  space" def| OR /regex literal/ ghi'
        expect.addSnapshotSerializer({
            serialize: result =>
                [
                    input,
                    ...(result.tokens as Token[])
                        .filter(token => result.sourceMap.has(token))
                        .map(token => annotateToken(token, '', result.sourceMap)),
                ].join('\n'),
            test: () => true,
        })

        expect(parse(input)).toMatchInlineSnapshot(`
                                 abc AND content:"with  space" def| OR /regex literal/ ghi
                                 ^^^ (pattern)
                                         ^^^^^^^ (filter.field: literal)
                                                 ^^^^^^^^^^^^^ (filter.value: literal)
                                         ^^^^^^^^^^^^^^^^^^^^^ (filter)
                                                               ^^^ (pattern)
                               `)
    })
})
