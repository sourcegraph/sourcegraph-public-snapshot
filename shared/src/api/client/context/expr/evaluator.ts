import { TokenType } from './lexer'
import { Expression, Parser, TemplateParser } from './parser'

/** A way to look up the value for an identifier. */
export interface ComputedContext {
    get(key: string): any
}

/** A computed context that returns undefined for every key. */
export const EMPTY_COMPUTED_CONTEXT: ComputedContext = {
    get: () => undefined,
}

/**
 * Evaluates an expression with the given context and returns the result.
 */
export function evaluate(expr: string, context: ComputedContext): any {
    return exec(new Parser().parse(expr), context)
}

/**
 * Evaluates a template with the given context and returns the result.
 *
 * A template is a string that interpolates expressions in ${...}. It uses the same syntax as
 * JavaScript templates.
 */
export function evaluateTemplate(template: string, context: ComputedContext): string {
    return exec(new TemplateParser().parse(template), context)
}

const FUNCS: { [name: string]: (...args: any[]) => any } = {
    get: (obj: any, key: string): any => obj[key],
    json: (obj: any): string => JSON.stringify(obj),
}

function exec(node: Expression, context: ComputedContext): any {
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
        const right = () => exec(node.Binary.right, context) // lazy evaluation
        switch (node.Binary.operator) {
            case '&&':
                return left && right()
            case '||':
                return left || right()
            case '==':
                // tslint:disable-next-line:triple-equals
                return left == right()
            case '!=':
                // tslint:disable-next-line:triple-equals
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
            return func.apply(null, args)
        }
        throw new SyntaxError(`Undefined function: ${expr.name}`)
    }

    throw new SyntaxError('Unrecognized syntax node')
}
