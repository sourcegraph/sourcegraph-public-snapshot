import assert from 'assert'
import { TokenType } from './lexer'
import { Expression, Parser, TemplateParser } from './parser'

describe('Parser', () => {
    const TESTS: { [expr: string]: Expression } = {
        '!a': {
            Unary: {
                operator: '!',
                expression: { Identifier: 'a' },
            },
        },
        '"a"': {
            Literal: {
                type: TokenType.String,
                value: 'a',
            },
        },
        '`a`': {
            Literal: {
                type: TokenType.String,
                value: 'a',
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        '`${x}`': {
            Template: {
                parts: [{ Identifier: 'x' }],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        '`${x}${y}`': {
            Template: {
                parts: [{ Identifier: 'x' }, { Identifier: 'y' }],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        '`${x+y}`': {
            Template: {
                parts: [
                    {
                        Binary: {
                            left: { Identifier: 'x' },
                            operator: '+',
                            right: { Identifier: 'y' },
                        },
                    },
                ],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        '`a${x}`': {
            Template: {
                parts: [{ Literal: { type: TokenType.String, value: 'a' } }, { Identifier: 'x' }],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        '`${x}b`': {
            Template: {
                parts: [{ Identifier: 'x' }, { Literal: { type: TokenType.String, value: 'b' } }],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        '`a${x}b`': {
            Template: {
                parts: [
                    { Literal: { type: TokenType.String, value: 'a' } },
                    { Identifier: 'x' },
                    { Literal: { type: TokenType.String, value: 'b' } },
                ],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        '`a${x}b${y}c`': {
            Template: {
                parts: [
                    { Literal: { type: TokenType.String, value: 'a' } },
                    { Identifier: 'x' },
                    { Literal: { type: TokenType.String, value: 'b' } },
                    { Identifier: 'y' },
                    { Literal: { type: TokenType.String, value: 'c' } },
                ],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        '`a${`x${y}z`}b`': {
            Template: {
                parts: [
                    { Literal: { type: TokenType.String, value: 'a' } },
                    {
                        Template: {
                            parts: [
                                { Literal: { type: TokenType.String, value: 'x' } },
                                { Identifier: 'y' },
                                { Literal: { type: TokenType.String, value: 'z' } },
                            ],
                        },
                    },
                    { Literal: { type: TokenType.String, value: 'b' } },
                ],
            },
        },
        'a && b': {
            Binary: {
                left: { Identifier: 'a' },
                operator: '&&',
                right: { Identifier: 'b' },
            },
        },
        '(a + b) * c': {
            Binary: {
                left: {
                    Binary: {
                        left: { Identifier: 'a' },
                        operator: '+',
                        right: { Identifier: 'b' },
                    },
                },
                operator: '*',
                right: { Identifier: 'c' },
            },
        },
        'ab * 1 + x(2, 3)': {
            Binary: {
                left: {
                    Binary: {
                        left: { Identifier: 'ab' },
                        operator: '*',
                        right: { Literal: { type: TokenType.Number, value: '1' } },
                    },
                },
                operator: '+',
                right: {
                    FunctionCall: {
                        name: 'x',
                        args: [
                            { Literal: { type: TokenType.Number, value: '2' } },
                            { Literal: { type: TokenType.Number, value: '3' } },
                        ],
                    },
                },
            },
        },
    }
    const parser = new Parser()
    for (const [expr, want] of Object.entries(TESTS)) {
        it(expr, () => assert.deepStrictEqual(parser.parse(expr), want))
    }

    it('throws an error on an invalid argument list', () => assert.throws(() => parser.parse('a(1,,)')))
    it('throws an error on an unclosed string', () => assert.throws(() => parser.parse('"')))
    it('throws an error on an unclosed template', () => assert.throws(() => parser.parse('`')))
    it('throws an error on an invalid unary operator', () => assert.throws(() => parser.parse('!')))
    it('throws an error on an invalid binary operator', () => assert.throws(() => parser.parse('a*')))
    it('throws an error on an unclosed function call', () => assert.throws(() => parser.parse('a(')))
    it('throws an error on an unterminated expression', () => assert.throws(() => parser.parse('(a=')))
})

describe('TemplateParser', () => {
    const TESTS: { [template: string]: Expression } = {
        // tslint:disable-next-line:no-invalid-template-strings
        '${x}': {
            Template: {
                parts: [{ Identifier: 'x' }],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        '${x}${y}': {
            Template: {
                parts: [{ Identifier: 'x' }, { Identifier: 'y' }],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        '${x+y}': {
            Template: {
                parts: [
                    {
                        Binary: {
                            left: { Identifier: 'x' },
                            operator: '+',
                            right: { Identifier: 'y' },
                        },
                    },
                ],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        'a${x}': {
            Template: {
                parts: [{ Literal: { type: TokenType.String, value: 'a' } }, { Identifier: 'x' }],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        '${x}b': {
            Template: {
                parts: [{ Identifier: 'x' }, { Literal: { type: TokenType.String, value: 'b' } }],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        'a${x}b': {
            Template: {
                parts: [
                    { Literal: { type: TokenType.String, value: 'a' } },
                    { Identifier: 'x' },
                    { Literal: { type: TokenType.String, value: 'b' } },
                ],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        'a${x}b${y}c': {
            Template: {
                parts: [
                    { Literal: { type: TokenType.String, value: 'a' } },
                    { Identifier: 'x' },
                    { Literal: { type: TokenType.String, value: 'b' } },
                    { Identifier: 'y' },
                    { Literal: { type: TokenType.String, value: 'c' } },
                ],
            },
        },
        // tslint:disable-next-line:no-invalid-template-strings
        'a${`x${y}z`}b': {
            Template: {
                parts: [
                    { Literal: { type: TokenType.String, value: 'a' } },
                    {
                        Template: {
                            parts: [
                                { Literal: { type: TokenType.String, value: 'x' } },
                                { Identifier: 'y' },
                                { Literal: { type: TokenType.String, value: 'z' } },
                            ],
                        },
                    },
                    { Literal: { type: TokenType.String, value: 'b' } },
                ],
            },
        },
    }
    const parser = new TemplateParser()
    for (const [template, want] of Object.entries(TESTS)) {
        it(template, () => assert.deepStrictEqual(parser.parse(template), want))
    }
})
