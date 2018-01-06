import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { ReplaySubject } from 'rxjs/ReplaySubject'

/**
 * Always represents the entire configuration cascade; i.e., it contains the
 * individual configs from the various config subjects (orgs, user, etc.).
 */
export const configurationCascade = new ReplaySubject<GQL.IConfigurationCascade>(1)

export interface SearchScopeConfiguration {
    name: string
    value: string
}

export interface SavedQueryConfiguration {
    description: string
    query?: string
    showOnHomepage?: boolean
}

export interface Configuration {
    ['search.scopes']?: SearchScopeConfiguration[]
    ['search.savedQueries']?: SavedQueryConfiguration[]
}

/**
 * Always represents the latest merged configuration for the current user
 * or visitor. Callers should cast the value to their own configuration type.
 */
export const currentConfiguration: Observable<Configuration> = configurationCascade.pipe(
    map(cascade => JSON.parse(cascade.merged.contents))
)
