import { TokenType } from './lexer'
import { ExpressionNode, Parser, TemplateParser } from './parser'

/**
 * A parsed context expression (that can evaluate to anything)
 */
export class Expression<T> {
    constructor(private root: ExpressionNode) {}

    public exec(context: ComputedContext): T {
        return exec(this.root, context)
    }
}

/**
 * A parsed template string that can contain `${contextKey}` interpolations.
 * Always evaluates to a string.
 */
export class TemplateExpression<S extends string = string> extends Expression<S> {}

/**
 * Evaluates an expression with the given context and returns the result.
 */
export function parse<T>(expr: string): Expression<T> {
    return new Expression<T>(new Parser().parse(expr))
}

/**
 * Evaluates a template with the given context and returns the result.
 *
 * A template is a string that interpolates expressions in ${...}. It uses the same syntax as
 * JavaScript templates.
 */
export function parseTemplate(template: string): TemplateExpression {
    return new TemplateExpression(new TemplateParser().parse(template))
}

/** A way to look up the value for an identifier. */
export interface ComputedContext {
    get(key: string): any
}

/** A computed context that returns undefined for every key. */
export const EMPTY_COMPUTED_CONTEXT: ComputedContext = {
    get: () => undefined,
}

const FUNCS: { [name: string]: (...args: any[]) => any } = {
    get: (obj: any, key: string): any => obj?.[key] ?? undefined,
    json: (obj: any): string => JSON.stringify(obj),
}

function exec(node: ExpressionNode, context: ComputedContext): any {
    if ('Literal' in node) {
        switch (node.Literal.type) {
            case TokenType.String:
                return node.Literal.value
            case TokenType.Number:
                return parseFloat(node.Literal.value)
            default:
                throw new SyntaxError(`Unrecognized literal of type ${TokenType[node.Literal.type]}`)
        }
    }

    if ('Template' in node) {
        const parts: any[] = []
        for (const expr of node.Template.parts) {
            parts.push(exec(expr, context))
        }
        return parts.join('')
    }

    if ('Binary' in node) {
        const left = exec(node.Binary.left, context)
        const right = (): any => exec(node.Binary.right, context) // lazy evaluation
        switch (node.Binary.operator) {
            case '&&':
                return left && right()
            case '||':
                return left || right()
            case '==':
                // eslint-disable-next-line eqeqeq
                return left == right()
            case '!=':
                // eslint-disable-next-line eqeqeq
                return left != right()
            case '===':
                return left === right()
            case '!==':
                return left !== right()
            case '<':
                return left < right()
            case '>':
                return left > right()
            case '<=':
                return left <= right()
            case '>=':
                return left >= right()
            case '+':
                return left + right()
            case '-':
                return left - right()
            case '*':
                return left * right()
            case '/':
                return left / right()
            case '^':
                return left ^ right()
            case '%':
                return left % right()
            default:
                throw new SyntaxError(`Invalid operator: ${node.Binary.operator}`)
        }
    }

    if ('Unary' in node) {
        const expr = exec(node.Unary.expression, context)
        switch (node.Unary.operator) {
            case '!':
                return !expr
            case '+':
                return expr
            case '-':
                return -expr
            default:
                throw new SyntaxError(`Invalid operator: ${node.Unary.operator}`)
        }
    }

    if ('Identifier' in node) {
        switch (node.Identifier) {
            case 'true':
                return true
            case 'false':
                return false
            case 'undefined':
                return undefined
            case 'null':
                return null
        }
        return context.get(node.Identifier)
    }

    if ('FunctionCall' in node) {
        const expr = node.FunctionCall
        const func = FUNCS[expr.name]
        if (typeof func === 'function') {
            const args = expr.args.map(arg => exec(arg, context))
            return func(...args)
        }
        throw new SyntaxError(`Undefined function: ${expr.name}`)
    }

    throw new SyntaxError('Unrecognized syntax node')
}
