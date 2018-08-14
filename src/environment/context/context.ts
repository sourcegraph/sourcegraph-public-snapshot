import { basename, dirname, extname } from 'path'
import { ConfigurationCascade } from '../../protocol'
import { Environment } from '../environment'
import { Extension } from '../extension'
import { evaluate } from './expr/evaluator'

/**
 * Context is an arbitrary, immutable set of key-value pairs.
 */
export interface Context {
    get(key: string): any
}

/**
 * A mutable context is a Context whose values can be set.
 */
export interface MutableContext extends Context {
    set(name: string, value: any): void
}

/**
 * Creates a new child context. All of the parent's values are available in the child context. Any values set on
 * the child context shadow parent context values. Setting the child context's values never modifies the parent
 * context (even if the parent is a MutableContext itself).
 */
export function createChildContext(parent: Context): MutableContext {
    const values = new Map<string, any>()
    return {
        get: (key: string): any => {
            const value = values.get(key)
            if (value !== undefined) {
                return value
            }
            return parent.get(key)
        },
        set: (name: string, value: any): void => {
            values.set(name, value)
        },
    }
}

/** A Context that is empty (and returns undefined for every key). */
export const EMPTY_CONTEXT: Context = {
    get(): any {
        return undefined
    },
}

/**
 * Wraps the environment's context, providing some context properties by deriving them from the
 * environment.
 *
 * @template X extension type
 * @template C configuration cascade type
 * @param next the context to use when the key doesn't refer to a property derived from the
 * environment
 */
export function environmentContext<X extends Extension, C extends ConfigurationCascade>(
    environment: Environment<X, C>,
    next: Context
): Context {
    return {
        get(key: string): any {
            if (key.startsWith('config.')) {
                const prop = key.slice('config.'.length)
                const value = environment.configuration.merged[prop]
                // Map undefined to null because an undefined value is treated as "does not exist in
                // context" and an error is thrown, which is undesirable for config values (for
                // which a falsey null default is useful).
                return value === undefined ? null : value
            }
            if (key === 'resource' || key === 'component') {
                return !!environment.component
            }
            if (key.startsWith('resource.')) {
                if (!environment.component) {
                    return undefined
                }
                const prop = key.slice('resource.'.length)
                switch (prop) {
                    case 'uri':
                        return environment.component.document.uri
                    case 'basename':
                        return basename(environment.component.document.uri)
                    case 'dirname':
                        return dirname(environment.component.document.uri)
                    case 'extname':
                        return extname(environment.component.document.uri)
                    case 'language':
                        return environment.component.document.languageId
                    case 'textContent':
                        return environment.component.document.text
                    case 'type':
                        return 'textDocument'
                }
            }
            if (key.startsWith('component.')) {
                if (!environment.component) {
                    return undefined
                }
                const prop = key.slice('component.'.length)
                switch (prop) {
                    case 'type':
                        return 'textEditor'
                }
            }
            return next.get(key)
        },
    }
}

/** Filters out items whose `when` context expression evaluates to false (or a falsey value). */
export function contextFilter<T extends { when?: string }>(context: Context, items: T[]): T[] {
    const keep: T[] = []
    for (const item of items) {
        if (item.when !== undefined) {
            if (!evaluate(item.when, createChildContext(context))) {
                continue // omit
            }
        }
        keep.push(item)
    }
    return keep
}
