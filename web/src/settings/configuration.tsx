import { ReplaySubject } from 'rxjs/ReplaySubject'

interface BaseConfiguration {
    [key: string]: string | number | boolean | (string | number | boolean | BaseConfiguration)[] | BaseConfiguration
}

/**
 * Always represents the latest merged configuration for the current user
 * or visitor. Callers should cast the value to their own configuration type.
 *
 * TODO(sqs): use a more sophisticated way of casting and getting component-specific
 * configuration, such as how VS Code sets defaults and ensures config sections are
 * set when you call (e.g.) getConfiguration.
 */
export const currentConfiguration = new ReplaySubject<BaseConfiguration>(1)
