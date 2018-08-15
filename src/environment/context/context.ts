import { basename, dirname, extname } from 'path'
import { Environment } from '../environment'

/**
 * Returns a new context created by applying the update context to the base context. It is equivalent to `{...base,
 * ...update}` in JavaScript except that null values in the update result in deletion of the property.
 */
export function applyContextUpdate(base: Context, update: Context): Context {
    const result = { ...base }
    for (const [key, value] of Object.entries(update)) {
        if (value === null) {
            delete result[key]
        } else {
            result[key] = value
        }
    }
    return result
}

/**
 * Context is an arbitrary, immutable set of key-value pairs.
 */
export interface Context {
    [key: string]: string | number | boolean | Context | null
}

/** A context that has no properties. */
export const EMPTY_CONTEXT: Context = {}

/**
 * Looks up a key in the computed context, which consists of special context properties (with higher precedence)
 * and the environment's context properties (with lower precedence). environment.
 *
 * @param key the context property key to look up
 */
export function getComputedContextProperty(environment: Environment, key: string): any {
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
        // TODO(sqs): Define these precisely. If the resource is in a repository, what is the "path"? Is it the
        // path relative to the repository's root? If it's a file on disk, then "path" could also mean the
        // (absolute) path on the file system. Clear up that ambiguity.
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
    if (key === 'context') {
        return environment.context
    }
    return environment.context[key]
}
