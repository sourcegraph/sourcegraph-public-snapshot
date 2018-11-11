import { Lexer, Operator, TokenType } from './lexer';
export declare type Expression = {
    FunctionCall: {
        name: string;
        args: Expression[];
    };
} | {
    Identifier: string;
} | {
    Literal: {
        type: TokenType.String | TokenType.Number;
        value: string;
    };
} | {
    Template: {
        parts: Expression[];
    };
} | {
    Unary: {
        operator: Operator;
        expression: Expression;
    };
} | {
    Binary: {
        operator: Operator;
        left: Expression;
        right: Expression;
    };
};
/**
 * Parses an expression.
 *
 * TODO: Operator precedence is not handled correctly. Use parentheses to be explicit about your desired
 * precedence.
 */
export declare class Parser {
    protected lexer: Lexer;
    parse(exprStr: string): Expression;
    private parseArgumentList;
    private parseFunctionCall;
    private parseTemplateParts;
    protected parseTemplate(): Expression;
    private parsePrimary;
    private parseUnary;
    private parseMultiplicative;
    private parseAdditive;
    private parseExpression;
}
/** Parses a template. */
export declare class TemplateParser extends Parser {
    parse(templateStr: string): Expression;
}
