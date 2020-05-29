import { Lexer, OPERATOR_CHARS, OPERATORS, OperatorTree, TemplateLexer, Token, TokenType } from './lexer'

describe('Lexer', () => {
    const l = new Lexer()

    test('scans an expression', () => {
        l.reset('ab * 2')
        expect(l.peek()).toEqual({ type: TokenType.Identifier, value: 'ab' } as Token)
        const token = l.next()
        expect(token).toEqual({ type: TokenType.Identifier, value: 'ab', start: 0, end: 2 } as Token)
        expect(l.next()).toBeTruthy()
        expect(l.next()).toBeTruthy()
        expect(l.peek()).toBe(undefined)
        expect(l.next()).toBe(undefined)

        l.reset('a')
        expect(l.next()).toEqual({ type: TokenType.Identifier, value: 'a', start: 0, end: 1 } as Token)
    })

    test('scans identifier with dots', () =>
        expect(scanAll(l, 'a.b')).toEqual([{ type: TokenType.Identifier, value: 'a.b', start: 0, end: 3 }] as Token[]))

    test('scans string', () =>
        expect(scanAll(l, '"a\\nb\\"c\'d"')).toEqual([
            { type: TokenType.String, value: 'a\nb"c\'d', start: 0, end: 11 },
        ] as Token[]))

    /* eslint-disable no-template-curly-in-string */
    describe('templates', () => {
        test('scans no-substitution template', () =>
            expect(scanAll(l, '`a`')).toEqual([
                { type: TokenType.NoSubstitutionTemplate, value: 'a', start: 0, end: 3 },
            ] as Token[]))

        test('scans template with empty head/tail', () =>
            expect(scanAll(l, '`${x}`')).toEqual([
                { type: TokenType.TemplateHead, value: '', start: 0, end: 3 },
                { type: TokenType.Identifier, value: 'x', start: 3, end: 4 },
                { type: TokenType.TemplateTail, value: '', start: 4, end: 6 },
            ] as Token[]))

        test('scans template with empty head/tail and multiple tokens', () =>
            expect(scanAll(l, '`${x+y}`')).toEqual([
                { type: TokenType.TemplateHead, value: '', start: 0, end: 3 },
                { type: TokenType.Identifier, value: 'x', start: 3, end: 4 },
                { type: TokenType.Operator, value: '+', start: 4, end: 5 },
                { type: TokenType.Identifier, value: 'y', start: 5, end: 6 },
                { type: TokenType.TemplateTail, value: '', start: 6, end: 8 },
            ] as Token[]))

        test('scans template with non-empty head, empty tail', () =>
            expect(scanAll(l, '`a${x}`')).toEqual([
                { type: TokenType.TemplateHead, value: 'a', start: 0, end: 4 },
                { type: TokenType.Identifier, value: 'x', start: 4, end: 5 },
                { type: TokenType.TemplateTail, value: '', start: 5, end: 7 },
            ] as Token[]))

        test('scans template with empty head, non-empty tail', () =>
            expect(scanAll(l, '`${x}b`')).toEqual([
                { type: TokenType.TemplateHead, value: '', start: 0, end: 3 },
                { type: TokenType.Identifier, value: 'x', start: 3, end: 4 },
                { type: TokenType.TemplateTail, value: 'b', start: 4, end: 7 },
            ] as Token[]))

        test('scans template with non-empty head/tail', () =>
            expect(scanAll(l, '`a${x}b`')).toEqual([
                { type: TokenType.TemplateHead, value: 'a', start: 0, end: 4 },
                { type: TokenType.Identifier, value: 'x', start: 4, end: 5 },
                { type: TokenType.TemplateTail, value: 'b', start: 5, end: 8 },
            ] as Token[]))

        test('scans template with middle and empty values', () =>
            expect(scanAll(l, '`${x}${y}`')).toEqual([
                { type: TokenType.TemplateHead, value: '', start: 0, end: 3 },
                { type: TokenType.Identifier, value: 'x', start: 3, end: 4 },
                { type: TokenType.TemplateMiddle, value: '', start: 4, end: 7 },
                { type: TokenType.Identifier, value: 'y', start: 7, end: 8 },
                { type: TokenType.TemplateTail, value: '', start: 8, end: 10 },
            ] as Token[]))

        test('scans template with middle', () =>
            expect(scanAll(l, '`a${x}b${y}c`')).toEqual([
                { type: TokenType.TemplateHead, value: 'a', start: 0, end: 4 },
                { type: TokenType.Identifier, value: 'x', start: 4, end: 5 },
                { type: TokenType.TemplateMiddle, value: 'b', start: 5, end: 9 },
                { type: TokenType.Identifier, value: 'y', start: 9, end: 10 },
                { type: TokenType.TemplateTail, value: 'c', start: 10, end: 13 },
            ] as Token[]))

        test('scans nested no-substitution template', () =>
            expect(scanAll(l, '`a${`x`}b`')).toEqual([
                { type: TokenType.TemplateHead, value: 'a', start: 0, end: 4 },
                { type: TokenType.NoSubstitutionTemplate, value: 'x', start: 4, end: 7 },
                { type: TokenType.TemplateTail, value: 'b', start: 7, end: 10 },
            ] as Token[]))

        test('scans nested template', () =>
            expect(scanAll(l, '`a${`x${y}z`}b`')).toEqual([
                { type: TokenType.TemplateHead, value: 'a', start: 0, end: 4 },
                { type: TokenType.TemplateHead, value: 'x', start: 4, end: 8 },
                { type: TokenType.Identifier, value: 'y', start: 8, end: 9 },
                { type: TokenType.TemplateTail, value: 'z', start: 9, end: 12 },
                { type: TokenType.TemplateTail, value: 'b', start: 12, end: 15 },
            ] as Token[]))

        test('throws on unclosed expression', () => expect(() => scanAll(l, 'x${')).toThrow())
    })
    /* eslint-enable no-template-curly-in-string */

    test('throws on unclosed string', () => {
        l.reset('"a')
        expect(() => l.next()).toThrow()
    })

    test('scans single-char binary operators', () =>
        expect(scanAll(l, 'a = b')).toEqual([
            { type: TokenType.Identifier, value: 'a', start: 0, end: 1 },
            { type: TokenType.Operator, value: '=', start: 2, end: 3 },
            { type: TokenType.Identifier, value: 'b', start: 4, end: 5 },
        ] as Token[]))

    test('scans 2-char binary operators', () =>
        expect(scanAll(l, 'a == b')).toEqual([
            { type: TokenType.Identifier, value: 'a', start: 0, end: 1 },
            { type: TokenType.Operator, value: '==', start: 2, end: 4 },
            { type: TokenType.Identifier, value: 'b', start: 5, end: 6 },
        ] as Token[]))

    test('scans 3-char binary operators', () =>
        expect(scanAll(l, 'a !== b')).toEqual([
            { type: TokenType.Identifier, value: 'a', start: 0, end: 1 },
            { type: TokenType.Operator, value: '!==', start: 2, end: 5 },
            { type: TokenType.Identifier, value: 'b', start: 6, end: 7 },
        ] as Token[]))

    test('scans adjacent operators', () => {
        expect(scanAll(l, 'a==!b')).toEqual([
            { type: TokenType.Identifier, value: 'a', start: 0, end: 1 },
            { type: TokenType.Operator, value: '==', start: 1, end: 3 },
            { type: TokenType.Operator, value: '!', start: 3, end: 4 },
            { type: TokenType.Identifier, value: 'b', start: 4, end: 5 },
        ] as Token[])
    })

    describe('constants', () => {
        test('OPERATOR_CHARS', () => {
            const maxOpLength = Math.max(...Object.keys(OPERATORS).map(s => s.length))
            for (const op of Object.keys(OPERATORS)) {
                if (op.length === 1) {
                    expect(
                        OPERATOR_CHARS[op] === true ||
                            (OPERATOR_CHARS[op] as { [ch: string]: OperatorTree })['\u0000'] === true
                    ).toBeTruthy()
                } else if (op.length === 2) {
                    expect(
                        (OPERATOR_CHARS[op.charAt(0)] as { [ch: string]: OperatorTree })[op.charAt(1)] === true ||
                            ((OPERATOR_CHARS[op.charAt(0)] as { [ch: string]: OperatorTree })[op.charAt(1)] as {
                                [ch: string]: OperatorTree
                            })['\u0000'] === true
                    ).toBeTruthy()
                } else if (op.length === 3) {
                    expect(
                        ((OPERATOR_CHARS[op.charAt(0)] as { [ch: string]: OperatorTree })[op.charAt(1)] as {
                            [ch: string]: OperatorTree
                        })[op.charAt(2)] === true ||
                            (((OPERATOR_CHARS[op.charAt(0)] as { [ch: string]: OperatorTree })[op.charAt(1)] as {
                                [ch: string]: OperatorTree
                            })[op.charAt(2)] as { [ch: string]: OperatorTree })['\u0000'] === true
                    ).toBeTruthy()
                } else if (op.length > maxOpLength) {
                    throw new Error(`operators of length ${op.length} are not yet supported`)
                }
            }
        })
    })
})

describe('TemplateLexer', () => {
    const l = new TemplateLexer()

    test('scans template with middle', () =>
        // eslint-disable-next-line no-template-curly-in-string
        expect(scanAll(l, 'a${x}b${y}c')).toEqual([
            { type: TokenType.TemplateHead, value: 'a', start: 0, end: 3 },
            { type: TokenType.Identifier, value: 'x', start: 3, end: 4 },
            { type: TokenType.TemplateMiddle, value: 'b', start: 4, end: 8 },
            { type: TokenType.Identifier, value: 'y', start: 8, end: 9 },
            { type: TokenType.TemplateTail, value: 'c', start: 9, end: 11 },
        ] as Token[]))
})

function scanAll(l: Lexer, exprStr: string): Token[] {
    const tokens: Token[] = []
    l.reset(exprStr)
    while (true) {
        const token = l.next()
        if (token) {
            tokens.push(token)
        } else {
            break
        }
    }
    return tokens
}
