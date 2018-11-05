import assert from 'assert'
import { Lexer, OPERATOR_CHARS, OPERATORS, OperatorTree, TemplateLexer, Token, TokenType } from './lexer'

describe('Lexer', () => {
    const l = new Lexer()

    it('scans an expression', () => {
        l.reset('ab * 2')
        assert.deepStrictEqual(l.peek(), { type: TokenType.Identifier, value: 'ab' } as Token)
        const token = l.next()
        assert.deepStrictEqual(token, { type: TokenType.Identifier, value: 'ab', start: 0, end: 2 } as Token)
        assert.ok(l.next())
        assert.ok(l.next())
        assert.strictEqual(l.peek(), undefined)
        assert.strictEqual(l.next(), undefined)

        l.reset('a')
        assert.deepStrictEqual(l.next(), { type: TokenType.Identifier, value: 'a', start: 0, end: 1 } as Token)
    })

    it('scans identifier with dots', () =>
        assert.deepStrictEqual(scanAll(l, 'a.b'), [
            { type: TokenType.Identifier, value: 'a.b', start: 0, end: 3 },
        ] as Token[]))

    it('scans string', () =>
        assert.deepStrictEqual(scanAll(l, '"a\\nb\\"c\'d"'), [
            { type: TokenType.String, value: 'a\nb"c\'d', start: 0, end: 11 },
        ] as Token[]))

    // tslint:disable:no-invalid-template-strings
    describe('templates', () => {
        it('scans no-substitution template', () =>
            assert.deepStrictEqual(scanAll(l, '`a`'), [
                { type: TokenType.NoSubstitutionTemplate, value: 'a', start: 0, end: 3 },
            ] as Token[]))

        it('scans template with empty head/tail', () =>
            assert.deepStrictEqual(scanAll(l, '`${x}`'), [
                { type: TokenType.TemplateHead, value: '', start: 0, end: 3 },
                { type: TokenType.Identifier, value: 'x', start: 3, end: 4 },
                { type: TokenType.TemplateTail, value: '', start: 4, end: 6 },
            ] as Token[]))

        it('scans template with empty head/tail and multiple tokens', () =>
            assert.deepStrictEqual(scanAll(l, '`${x+y}`'), [
                { type: TokenType.TemplateHead, value: '', start: 0, end: 3 },
                { type: TokenType.Identifier, value: 'x', start: 3, end: 4 },
                { type: TokenType.Operator, value: '+', start: 4, end: 5 },
                { type: TokenType.Identifier, value: 'y', start: 5, end: 6 },
                { type: TokenType.TemplateTail, value: '', start: 6, end: 8 },
            ] as Token[]))

        it('scans template with non-empty head, empty tail', () =>
            assert.deepStrictEqual(scanAll(l, '`a${x}`'), [
                { type: TokenType.TemplateHead, value: 'a', start: 0, end: 4 },
                { type: TokenType.Identifier, value: 'x', start: 4, end: 5 },
                { type: TokenType.TemplateTail, value: '', start: 5, end: 7 },
            ] as Token[]))

        it('scans template with empty head, non-empty tail', () =>
            assert.deepStrictEqual(scanAll(l, '`${x}b`'), [
                { type: TokenType.TemplateHead, value: '', start: 0, end: 3 },
                { type: TokenType.Identifier, value: 'x', start: 3, end: 4 },
                { type: TokenType.TemplateTail, value: 'b', start: 4, end: 7 },
            ] as Token[]))

        it('scans template with non-empty head/tail', () =>
            assert.deepStrictEqual(scanAll(l, '`a${x}b`'), [
                { type: TokenType.TemplateHead, value: 'a', start: 0, end: 4 },
                { type: TokenType.Identifier, value: 'x', start: 4, end: 5 },
                { type: TokenType.TemplateTail, value: 'b', start: 5, end: 8 },
            ] as Token[]))

        it('scans template with middle and empty values', () =>
            assert.deepStrictEqual(scanAll(l, '`${x}${y}`'), [
                { type: TokenType.TemplateHead, value: '', start: 0, end: 3 },
                { type: TokenType.Identifier, value: 'x', start: 3, end: 4 },
                { type: TokenType.TemplateMiddle, value: '', start: 4, end: 7 },
                { type: TokenType.Identifier, value: 'y', start: 7, end: 8 },
                { type: TokenType.TemplateTail, value: '', start: 8, end: 10 },
            ] as Token[]))

        it('scans template with middle', () =>
            assert.deepStrictEqual(scanAll(l, '`a${x}b${y}c`'), [
                { type: TokenType.TemplateHead, value: 'a', start: 0, end: 4 },
                { type: TokenType.Identifier, value: 'x', start: 4, end: 5 },
                { type: TokenType.TemplateMiddle, value: 'b', start: 5, end: 9 },
                { type: TokenType.Identifier, value: 'y', start: 9, end: 10 },
                { type: TokenType.TemplateTail, value: 'c', start: 10, end: 13 },
            ] as Token[]))

        it('scans nested no-substitution template', () =>
            assert.deepStrictEqual(scanAll(l, '`a${`x`}b`'), [
                { type: TokenType.TemplateHead, value: 'a', start: 0, end: 4 },
                { type: TokenType.NoSubstitutionTemplate, value: 'x', start: 4, end: 7 },
                { type: TokenType.TemplateTail, value: 'b', start: 7, end: 10 },
            ] as Token[]))

        it('scans nested template', () =>
            assert.deepStrictEqual(scanAll(l, '`a${`x${y}z`}b`'), [
                { type: TokenType.TemplateHead, value: 'a', start: 0, end: 4 },
                { type: TokenType.TemplateHead, value: 'x', start: 4, end: 8 },
                { type: TokenType.Identifier, value: 'y', start: 8, end: 9 },
                { type: TokenType.TemplateTail, value: 'z', start: 9, end: 12 },
                { type: TokenType.TemplateTail, value: 'b', start: 12, end: 15 },
            ] as Token[]))

        it('throws on unclosed expression', () => assert.throws(() => scanAll(l, 'x${')))
    })
    // tslint:enable:no-invalid-template-strings

    it('throws on unclosed string', () => {
        l.reset('"a')
        assert.throws(() => l.next())
    })

    it('scans single-char binary operators', () =>
        assert.deepStrictEqual(scanAll(l, 'a = b'), [
            { type: TokenType.Identifier, value: 'a', start: 0, end: 1 },
            { type: TokenType.Operator, value: '=', start: 2, end: 3 },
            { type: TokenType.Identifier, value: 'b', start: 4, end: 5 },
        ] as Token[]))

    it('scans 2-char binary operators', () =>
        assert.deepStrictEqual(scanAll(l, 'a == b'), [
            { type: TokenType.Identifier, value: 'a', start: 0, end: 1 },
            { type: TokenType.Operator, value: '==', start: 2, end: 4 },
            { type: TokenType.Identifier, value: 'b', start: 5, end: 6 },
        ] as Token[]))

    it('scans 3-char binary operators', () =>
        assert.deepStrictEqual(scanAll(l, 'a !== b'), [
            { type: TokenType.Identifier, value: 'a', start: 0, end: 1 },
            { type: TokenType.Operator, value: '!==', start: 2, end: 5 },
            { type: TokenType.Identifier, value: 'b', start: 6, end: 7 },
        ] as Token[]))

    it('scans adjacent operators', () => {
        assert.deepStrictEqual(scanAll(l, 'a==!b'), [
            { type: TokenType.Identifier, value: 'a', start: 0, end: 1 },
            { type: TokenType.Operator, value: '==', start: 1, end: 3 },
            { type: TokenType.Operator, value: '!', start: 3, end: 4 },
            { type: TokenType.Identifier, value: 'b', start: 4, end: 5 },
        ] as Token[])
    })

    describe('constants', () => {
        it('OPERATOR_CHARS', () => {
            const maxOpLength = Math.max(...Object.keys(OPERATORS).map(s => s.length))
            for (const op of Object.keys(OPERATORS)) {
                if (op.length === 1) {
                    assert.ok(
                        OPERATOR_CHARS[op] === true ||
                            (OPERATOR_CHARS[op] as { [ch: string]: OperatorTree })['\x00'] === true,
                        op
                    )
                } else if (op.length === 2) {
                    assert.ok(
                        (OPERATOR_CHARS[op.charAt(0)] as { [ch: string]: OperatorTree })[op.charAt(1)] === true ||
                            ((OPERATOR_CHARS[op.charAt(0)] as { [ch: string]: OperatorTree })[op.charAt(1)] as {
                                [ch: string]: OperatorTree
                            })['\x00'] === true,
                        op
                    )
                } else if (op.length === 3) {
                    assert.ok(
                        ((OPERATOR_CHARS[op.charAt(0)] as { [ch: string]: OperatorTree })[op.charAt(1)] as {
                            [ch: string]: OperatorTree
                        })[op.charAt(2)] === true ||
                            (((OPERATOR_CHARS[op.charAt(0)] as { [ch: string]: OperatorTree })[op.charAt(1)] as {
                                [ch: string]: OperatorTree
                            })[op.charAt(2)] as { [ch: string]: OperatorTree })['\x00'] === true,
                        op
                    )
                } else if (op.length > maxOpLength) {
                    throw new Error(`operators of length ${op.length} are not yet supported`)
                }
            }
        })
    })
})

describe('TemplateLexer', () => {
    const l = new TemplateLexer()

    it('scans template with middle', () =>
        // tslint:disable-next-line:no-invalid-template-strings
        assert.deepStrictEqual(scanAll(l, 'a${x}b${y}c'), [
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
