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
     * ECMAScript lexical components for templates, upon which this is based.
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
    /** The type of this token. */
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

/** The token that indicates the beginning of a template string. */
export const TEMPLATE_BEGIN = '${' // tslint:disable-line:no-invalid-template-strings

/** Scans an expression. */
export class Lexer {
    private expression = ''
    private length = 0
    protected _index = 0
    private marker = 0
    protected curlyStack = 0

    /** The current character position of the lexer's cursor. */
    public get index(): number {
        return this._index
    }

    public reset(str: string): void {
        this.expression = str
        this.length = str.length
        this._index = 0
        this.curlyStack = 0
    }

    public next(): Token | undefined {
        this.skipSpaces()
        if (this._index >= this.length) {
            return undefined
        }

        this.marker = this._index

        const token = this.scanNext()
        if (token !== undefined) {
            return token
        }

        throw new SyntaxError(`Unexpected character ${JSON.stringify(this.peekNextChar())} (at ${this.index})`)
    }

    public peek(): Pick<Token, Exclude<keyof Token, 'start' | 'end'>> | undefined {
        const savedIndex = this._index
        const savedCurlyStack = this.curlyStack
        let token: Token | undefined
        try {
            token = this.next()
            if (token) {
                delete token.start
                delete token.end
            }
        } catch (e) {
            token = undefined
        }
        this._index = savedIndex
        this.curlyStack = savedCurlyStack

        return token
    }

    protected scanNext(): Token | undefined {
        let token = this.scanString()
        if (token !== undefined) {
            return token
        }

        token = this.scanTemplate()
        if (token !== undefined) {
            return token
        }

        token = this.scanNumber()
        if (token !== undefined) {
            return token
        }

        token = this.scanOperator()
        if (token !== undefined) {
            return token
        }

        token = this.scanIdentifier()
        if (token !== undefined) {
            return token
        }

        return undefined
    }

    private peekNextChar(advance = 0): string {
        const idx = this._index + advance
        return idx < this.length ? this.expression.charAt(idx) : '\x00'
    }

    private getNextChar(): string {
        let ch = '\x00'
        const idx = this._index
        if (idx < this.length) {
            ch = this.expression.charAt(idx)
            this._index += 1
        }
        return ch
    }

    private createToken(type: TokenType, value: any): Token {
        return {
            type,
            value,
            start: this.marker,
            end: this._index,
        }
    }

    private skipSpaces(): void {
        while (this._index < this.length) {
            const ch = this.peekNextChar()
            if (!isWhiteSpace(ch)) {
                break
            }
            this.getNextChar()
        }
    }

    private scanOperator(): Token | undefined {
        let searchTree: OperatorTree | boolean = OPERATOR_CHARS
        let value = ''
        while (searchTree && searchTree !== true) {
            const ch = this.peekNextChar()
            searchTree = searchTree[ch]
            if (searchTree) {
                value += ch
                this.getNextChar()
            }
        }
        if (value === '}') {
            this.curlyStack--
        }
        if (value === '') {
            return undefined
        }
        return this.createToken(TokenType.Operator, value)
    }

    private scanIdentifier(): Token | undefined {
        let ch = this.peekNextChar()
        if (!isIdentifierStart(ch)) {
            return undefined
        }

        let id = this.getNextChar()
        while (true) {
            ch = this.peekNextChar()
            if (!isIdentifierPart(ch)) {
                break
            }
            id += this.getNextChar()
        }

        return this.createToken(TokenType.Identifier, id)
    }

    private scanString(): Token | undefined {
        const quote = this.peekNextChar()
        if (quote !== "'" && quote !== '"') {
            return undefined
        }
        this.getNextChar()

        let terminated = false
        let str = ''
        while (this._index < this.length) {
            const ch = this.getNextChar()
            if (ch === quote) {
                terminated = true
                break
            }
            if (ch === '\\') {
                str += backslashEscapeCodeString(this.getNextChar())
            } else {
                str += ch
            }
        }
        if (!terminated) {
            throw new Error(`Unterminated string literal (at ${this.index})`)
        }
        return this.createToken(TokenType.String, str)
    }

    private scanTemplate(): Token | undefined {
        const ch = this.peekNextChar()
        // tslint:disable-next-line:no-invalid-template-strings
        if (!(ch === '`' || (ch === '}' && this.curlyStack > 0))) {
            return undefined
        }
        this.getNextChar()

        const head = ch === '`'
        return this.doScanTemplate(head)
    }

    protected backtick(): boolean {
        return true
    }

    protected doScanTemplate(head: boolean): Token {
        let tail = false

        let terminated = false
        let hasSubstitution = false
        let str = ''
        while (this._index < this.length) {
            const ch = this.getNextChar()
            if (ch === '`' && this.backtick()) {
                tail = true
                terminated = true
                break
            }
            if (ch === '\\') {
                str += backslashEscapeCodeString(this.getNextChar())
            } else {
                if (ch === '$') {
                    const ch2 = this.peekNextChar()
                    if (ch2 === '{') {
                        // tslint:disable-next-line:no-invalid-template-strings
                        this.curlyStack++
                        this.getNextChar()
                        terminated = true
                        hasSubstitution = true
                        break
                    }
                }
                str += ch
            }
        }
        if (!head) {
            this.curlyStack--
        }
        if (this.backtick()) {
            if (!terminated) {
                throw new Error(`Unterminated template literal (at ${this.index})`)
            }
        } else if (this._index === this.length) {
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
        return this.createToken(type, str)
    }

    private scanNumber(): Token | undefined {
        let ch = this.peekNextChar()
        if (!isDecimalDigit(ch) && ch !== '.') {
            return undefined
        }

        let num = ''
        if (ch !== '.') {
            num = this.getNextChar()
            while (true) {
                ch = this.peekNextChar()
                if (!isDecimalDigit(ch)) {
                    break
                }
                num += this.getNextChar()
            }
        }

        if (ch === '.') {
            num += this.getNextChar()
            while (true) {
                ch = this.peekNextChar()
                if (!isDecimalDigit(ch)) {
                    break
                }
                num += this.getNextChar()
            }
        }

        if (ch === 'e' || ch === 'E') {
            num += this.getNextChar()
            ch = this.peekNextChar()
            if (ch === '+' || ch === '-' || isDecimalDigit(ch)) {
                num += this.getNextChar()
                while (true) {
                    ch = this.peekNextChar()
                    if (!isDecimalDigit(ch)) {
                        break
                    }
                    num += this.getNextChar()
                }
            } else {
                ch = `character ${JSON.stringify(ch)}`
                if (this._index >= this.length) {
                    ch = '<end>'
                }
                throw new SyntaxError(`Unexpected ${ch} after the exponent sign (at ${this.index})`)
            }
        }

        if (num === '.') {
            throw new SyntaxError(`Expected decimal digits after the dot sign (at ${this.index})`)
        }

        return this.createToken(TokenType.Number, num)
    }
}

/** Scans a template. */
export class TemplateLexer extends Lexer {
    public next(): Token | undefined {
        if (this._index === 0) {
            return this.doScanTemplate(true)
        }
        return super.next()
    }

    protected backtick(): boolean {
        // The root is not surrounded with backticks.
        return this.curlyStack !== 0
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
