import { Environment } from '../environment';
import { TextDocumentItem } from '../types/textDocument';
/**
 * Returns a new context created by applying the update context to the base context. It is equivalent to `{...base,
 * ...update}` in JavaScript except that null values in the update result in deletion of the property.
 */
export declare function applyContextUpdate(base: Context, update: Context): Context;
/**
 * Context is an arbitrary, immutable set of key-value pairs.
 */
export interface Context {
    [key: string]: string | number | boolean | Context | null;
}
/** A context that has no properties. */
export declare const EMPTY_CONTEXT: Context;
/**
 * Looks up a key in the computed context, which consists of special context properties (with higher precedence)
 * and the environment's context properties (with lower precedence).
 *
 * @param key the context property key to look up
 * @param scope the user interface component in whose scope this computation should occur
 */
export declare function getComputedContextProperty(environment: Environment, key: string, scope?: TextDocumentItem): any;
