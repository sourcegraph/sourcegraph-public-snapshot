/** Valid token types in expressions. */
export declare enum TokenType {
    /** An operator. */
    Operator = 0,
    /** An identifier. */
    Identifier = 1,
    /** A string literal. */
    String = 2,
    /**
     * The start of a template until its first expression.
     *
     * See https://tc39.github.io/ecma262/#sec-template-literal-lexical-components for documentation on the
     * ECMAScript lexical components for templates, upon which this is based.
     */
    TemplateHead = 3,
    /** The end of a previous template expression until the next template expression. */
    TemplateMiddle = 4,
    /** The end of a previous template expression until the end of the template. */
    TemplateTail = 5,
    /** A template with no substitutions. */
    NoSubstitutionTemplate = 6,
    /** A number literal. */
    Number = 7
}
/** A token that the expression lexer scanned in an expression. */
export interface Token {
    /** The type of this token. */
    type: TokenType;
    /**
     * The token's value.
     *
     * For string and template literals, this is the parsed string value (after accounting for escape sequences but
     * not template expressions). For number literals, this is the (unparsed) string representation.
     */
    value: any;
    /** The start character position of this token. */
    start: number;
    /** The end character position of this token. */
    end: number;
}
/**
 * All valid operators in expressions. The values are the operator precedence (or, for operators that are not operators, 0). This
 * must be kept in sync with OPERATOR_CHARS.
 *
 * Exported for testing only.
 */
export declare const OPERATORS: {
    '(': number;
    ')': number;
    '}': number;
    ',': number;
    '=': number;
    '||': number;
    '&&': number;
    '^': number;
    '==': number;
    '!=': number;
    '===': number;
    '!==': number;
    '<': number;
    '>': number;
    '<=': number;
    '>=': number;
    '+': number;
    '-': number;
    '*': number;
    '/': number;
    '%': number;
    '!': number;
};
/** All valid operators. */
export declare type Operator = keyof typeof OPERATORS;
export declare type OperatorTree = boolean | {
    [ch: string]: OperatorTree;
};
/**
 * A tree with the next valid operator characters for multi-character operators. This must be kept in sync with
 * OPERATORS.
 *
 * Exported for testing only.
 */
export declare const OPERATOR_CHARS: {
    [ch: string]: OperatorTree;
};
/** The token that indicates the beginning of a template string. */
export declare const TEMPLATE_BEGIN = "${";
/** Scans an expression. */
export declare class Lexer {
    private expression;
    private length;
    protected _index: number;
    private marker;
    protected curlyStack: number;
    /** The current character position of the lexer's cursor. */
    readonly index: number;
    reset(str: string): void;
    next(): Token | undefined;
    peek(): Pick<Token, Exclude<keyof Token, 'start' | 'end'>> | undefined;
    protected scanNext(): Token | undefined;
    private peekNextChar;
    private getNextChar;
    private createToken;
    private skipSpaces;
    private scanOperator;
    private scanIdentifier;
    private scanString;
    private scanTemplate;
    protected backtick(): boolean;
    protected doScanTemplate(head: boolean): Token;
    private scanNumber;
}
/** Scans a template. */
export declare class TemplateLexer extends Lexer {
    next(): Token | undefined;
    protected backtick(): boolean;
}
