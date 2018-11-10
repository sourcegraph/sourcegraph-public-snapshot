/** A way to look up the value for an identifier. */
export interface ComputedContext {
    get(key: string): any;
}
/** A computed context that returns undefined for every key. */
export declare const EMPTY_COMPUTED_CONTEXT: ComputedContext;
/**
 * Evaluates an expression with the given context and returns the result.
 */
export declare function evaluate(expr: string, context: ComputedContext): any;
/**
 * Evaluates a template with the given context and returns the result.
 *
 * A template is a string that interpolates expressions in ${...}. It uses the same syntax as
 * JavaScript templates.
 */
export declare function evaluateTemplate(template: string, context: ComputedContext): any;
