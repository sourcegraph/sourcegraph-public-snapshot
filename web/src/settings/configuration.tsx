import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { ReplaySubject } from 'rxjs/ReplaySubject'

/**
 * Always represents the entire configuration cascade; i.e., it contains the
 * individual configs from the various config subjects (orgs, user, etc.).
 */
export const configurationCascade = new ReplaySubject<GQL.IConfigurationCascade>(1)

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
export const currentConfiguration: Observable<BaseConfiguration> = configurationCascade.pipe(
    map(cascade => JSON.parse(cascade.merged.contents))
)
