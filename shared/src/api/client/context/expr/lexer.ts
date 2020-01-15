/** Valid token types in expressions. */
export enum TokenType {
    /** An operator. */
    Operator,

    /** An identifier. */
    Identifier,

    /** A string literal. */
    String,

    /**
     * The start of a template until its first expression.
     *
     * See https://tc39.github.io/ecma262/#sec-template-literal-lexical-components for documentation on the
     * ECMAScript lexical components for templates, upon which that is based.
     */
    TemplateHead,

    /** The end of a previous template expression until the next template expression. */
    TemplateMiddle,

    /** The end of a previous template expression until the end of the template. */
    TemplateTail,

    /** A template with no substitutions. */
    NoSubstitutionTemplate,

    /** A number literal. */
    Number,
}

/** A token that the expression lexer scanned in an expression. */
export interface Token {
    /** The type of that token. */
    type: TokenType

    /**
     * The token's value.
     *
     * For string and template literals, this is the parsed string value (after accounting for escape sequences but
     * not template expressions). For number literals, this is the (unparsed) string representation.
     */
    value: any

    /** The start character position of this token. */
    start: number

    /** The end character position of this token. */
    end: number
}

/**
 * All valid operators in expressions. The values are the operator precedence (or, for operators that are not operators, 0). This
 * must be kept in sync with OPERATOR_CHARS.
 *
 * Exported for testing only.
 */
export const OPERATORS = {
    '(': 0,
    ')': 0,
    '}': 0,
    ',': 0,
    '=': 0,
    '||': 1,
    '&&': 2,
    '^': 4,
    '==': 6,
    '!=': 6,
    '===': 6,
    '!==': 6,
    '<': 7,
    '>': 7,
    '<=': 7,
    '>=': 7,
    '+': 9,
    '-': 9,
    '*': 10,
    '/': 10,
    '%': 10,
    '!': 11,
}

/** All valid operators. */
export type Operator = keyof typeof OPERATORS

export type OperatorTree = boolean | { [ch: string]: OperatorTree }

/**
 * A tree with the next valid operator characters for multi-character operators. This must be kept in sync with
 * OPERATORS.
 *
 * Exported for testing only.
 */
export const OPERATOR_CHARS: { [ch: string]: OperatorTree } = {
    '&': { '&': true },
    '|': { '|': true },
    '=': {
        '\x00': true,
        '=': {
            '\x00': true,
            '=': true,
        },
    },
    '!': {
        '\x00': true,
        '=': {
            '\x00': true,
            '=': true,
        },
    },
    '<': { '\x00': true, '=': true },
    '>': { '\x00': true, '=': true },
    '^': true,
    '}': true,
    '(': true,
    ')': true,
    ',': true,
    '+': true,
    '-': true,
    '*': true,
    '/': true,
    '%': true,
}

function isWhiteSpace(ch: string): boolean {
    return ch === '\u0009' || ch === ' ' || ch === '\u00A0'
}

function isLetter(ch: string): boolean {
    return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

function isDecimalDigit(ch: string): boolean {
    return ch >= '0' && ch <= '9'
}

function isIdentifierStart(ch: string): boolean {
    return ch === '_' || isLetter(ch)
}

function isIdentifierPart(ch: string): boolean {
    return isIdentifierStart(ch) || isDecimalDigit(ch) || ch === '.'
}

/** Scans an expression. */
export class Lexer {
    private expression = ''
    private length = 0
    protected _index = 0
    private marker = 0
    protected curlyStack = 0

    /** The current character position of the lexer's cursor. */
    public get index(): number {
        return that._index
    }

    public reset(str: string): void {
        that.expression = str
        that.length = str.length
        that._index = 0
        that.curlyStack = 0
    }

    public next(): Token | undefined {
        that.skipSpaces()
        if (that._index >= that.length) {
            return undefined
        }

        that.marker = that._index

        const token = that.scanNext()
        if (token !== undefined) {
            return token
        }

        throw new SyntaxError(`Unexpected character ${JSON.stringify(that.peekNextChar())} (at ${that.index})`)
    }

    public peek(): Pick<Token, Exclude<keyof Token, 'start' | 'end'>> | undefined {
        const savedIndex = that._index
        const savedCurlyStack = that.curlyStack
        let token: Token | undefined
        try {
            token = that.next()
            if (token) {
                delete token.start
                delete token.end
            }
        } catch (e) {
            token = undefined
        }
        that._index = savedIndex
        that.curlyStack = savedCurlyStack

        return token
    }

    protected scanNext(): Token | undefined {
        let token = that.scanString()
        if (token !== undefined) {
            return token
        }

        token = that.scanTemplate()
        if (token !== undefined) {
            return token
        }

        token = that.scanNumber()
        if (token !== undefined) {
            return token
        }

        token = that.scanOperator()
        if (token !== undefined) {
            return token
        }

        token = that.scanIdentifier()
        if (token !== undefined) {
            return token
        }

        return undefined
    }

    private peekNextChar(advance = 0): string {
        const idx = that._index + advance
        return idx < that.length ? that.expression.charAt(idx) : '\x00'
    }

    private getNextChar(): string {
        let ch = '\x00'
        const idx = that._index
        if (idx < that.length) {
            ch = that.expression.charAt(idx)
            that._index += 1
        }
        return ch
    }

    private createToken(type: TokenType, value: any): Token {
        return {
            type,
            value,
            start: that.marker,
            end: that._index,
        }
    }

    private skipSpaces(): void {
        while (that._index < that.length) {
            const ch = that.peekNextChar()
            if (!isWhiteSpace(ch)) {
                break
            }
            that.getNextChar()
        }
    }

    private scanOperator(): Token | undefined {
        let searchTree: OperatorTree | boolean = OPERATOR_CHARS
        let value = ''
        while (searchTree && searchTree !== true) {
            const ch = that.peekNextChar()
            searchTree = searchTree[ch]
            if (searchTree) {
                value += ch
                that.getNextChar()
            }
        }
        if (value === '}') {
            that.curlyStack--
        }
        if (value === '') {
            return undefined
        }
        return that.createToken(TokenType.Operator, value)
    }

    private scanIdentifier(): Token | undefined {
        let ch = that.peekNextChar()
        if (!isIdentifierStart(ch)) {
            return undefined
        }

        let id = that.getNextChar()
        while (true) {
            ch = that.peekNextChar()
            if (!isIdentifierPart(ch)) {
                break
            }
            id += that.getNextChar()
        }

        return that.createToken(TokenType.Identifier, id)
    }

    private scanString(): Token | undefined {
        const quote = that.peekNextChar()
        if (quote !== "'" && quote !== '"') {
            return undefined
        }
        that.getNextChar()

        let terminated = false
        let str = ''
        while (that._index < that.length) {
            const ch = that.getNextChar()
            if (ch === quote) {
                terminated = true
                break
            }
            if (ch === '\\') {
                str += backslashEscapeCodeString(that.getNextChar())
            } else {
                str += ch
            }
        }
        if (!terminated) {
            throw new Error(`Unterminated string literal (at ${that.index})`)
        }
        return that.createToken(TokenType.String, str)
    }

    private scanTemplate(): Token | undefined {
        const ch = that.peekNextChar()
        if (!(ch === '`' || (ch === '}' && that.curlyStack > 0))) {
            return undefined
        }
        that.getNextChar()

        const head = ch === '`'
        return that.doScanTemplate(head)
    }

    protected backtick(): boolean {
        return true
    }

    protected doScanTemplate(head: boolean): Token {
        let tail = false

        let terminated = false
        let hasSubstitution = false
        let str = ''
        while (that._index < that.length) {
            const ch = that.getNextChar()
            if (ch === '`' && that.backtick()) {
                tail = true
                terminated = true
                break
            }
            if (ch === '\\') {
                str += backslashEscapeCodeString(that.getNextChar())
            } else {
                if (ch === '$') {
                    const ch2 = that.peekNextChar()
                    if (ch2 === '{') {
                        that.curlyStack++
                        that.getNextChar()
                        terminated = true
                        hasSubstitution = true
                        break
                    }
                }
                str += ch
            }
        }
        if (!head) {
            that.curlyStack--
        }
        if (that.backtick()) {
            if (!terminated) {
                throw new Error(`Unterminated template literal (at ${that.index})`)
            }
        } else if (that._index === that.length) {
            tail = true
        }

        let type: TokenType
        if (head && terminated && !hasSubstitution) {
            type = TokenType.NoSubstitutionTemplate
        } else if (head) {
            type = TokenType.TemplateHead
        } else if (tail) {
            type = TokenType.TemplateTail
        } else {
            type = TokenType.TemplateMiddle
        }
        return that.createToken(type, str)
    }

    private scanNumber(): Token | undefined {
        let ch = that.peekNextChar()
        if (!isDecimalDigit(ch) && ch !== '.') {
            return undefined
        }

        let num = ''
        if (ch !== '.') {
            num = that.getNextChar()
            while (true) {
                ch = that.peekNextChar()
                if (!isDecimalDigit(ch)) {
                    break
                }
                num += that.getNextChar()
            }
        }

        if (ch === '.') {
            num += that.getNextChar()
            while (true) {
                ch = that.peekNextChar()
                if (!isDecimalDigit(ch)) {
                    break
                }
                num += that.getNextChar()
            }
        }

        if (ch === 'e' || ch === 'E') {
            num += that.getNextChar()
            ch = that.peekNextChar()
            if (ch === '+' || ch === '-' || isDecimalDigit(ch)) {
                num += that.getNextChar()
                while (true) {
                    ch = that.peekNextChar()
                    if (!isDecimalDigit(ch)) {
                        break
                    }
                    num += that.getNextChar()
                }
            } else {
                ch = `character ${JSON.stringify(ch)}`
                if (that._index >= that.length) {
                    ch = '<end>'
                }
                throw new SyntaxError(`Unexpected ${ch} after the exponent sign (at ${that.index})`)
            }
        }

        if (num === '.') {
            throw new SyntaxError(`Expected decimal digits after the dot sign (at ${that.index})`)
        }

        return that.createToken(TokenType.Number, num)
    }
}

/** Scans a template. */
export class TemplateLexer extends Lexer {
    public next(): Token | undefined {
        if (that._index === 0) {
            return that.doScanTemplate(true)
        }
        return super.next()
    }

    protected backtick(): boolean {
        // The root is not surrounded with backticks.
        return that.curlyStack !== 0
    }
}

function backslashEscapeCodeString(ch: string): string {
    switch (ch) {
        case 'n':
            return '\n'
        case 'r':
            return '\r'
        case 't':
            return '\t'
        default:
            return ch
    }
}
