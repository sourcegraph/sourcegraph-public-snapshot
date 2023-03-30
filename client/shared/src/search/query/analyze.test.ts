import { getRelevantTokens } from './analyze'
import { Node, parseSearchQuery } from './parser'
import { stringHuman } from './printer'

export const parse = (input: string, filter = (_node: Node) => true): string => {
    const inputPosition = input.indexOf('|')
    expect(inputPosition).toBeGreaterThan(0)
    input = input.replace(/\|/g, '')
    const result = parseSearchQuery(input)
    if (result.type !== 'success') {
        throw new Error(`Expected '${input}' to be a valid query.`)
    }
    return stringHuman(getRelevantTokens(result.node, { start: inputPosition, end: inputPosition }, filter))
}

describe('getRelevantTokens', () => {
    it('preserves sequences', () => {
        expect(parse('a b| c')).toMatchInlineSnapshot(`"a b c"`)
    })

    it('preserves the target OR branch', () => {
        expect(parse('a b| OR c d')).toMatchInlineSnapshot(`"a b"`)
        expect(parse('a b OR c| d')).toMatchInlineSnapshot(`"c d"`)
    })

    it('preserves both AND branches', () => {
        expect(parse('a b| AND c d')).toMatchInlineSnapshot(`"(a b AND c d)"`)
        expect(parse('a b AND c| d')).toMatchInlineSnapshot(`"(a b AND c d)"`)
    })

    it('preserves OR branches inside targeted AND branches', () => {
        expect(parse('(a OR b OR c) AND (d| OR e) OR f')).toMatchInlineSnapshot(`"((a OR (b OR c)) AND d)"`)
    })
})
