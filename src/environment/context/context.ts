import { basename, dirname, extname } from 'path'
import { Environment } from '../environment'

/**
 * A context is an immutable map of keys to values.
 */
export interface Context {
    [key: string]: string | number | boolean
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
    return environment.context[key]
}
