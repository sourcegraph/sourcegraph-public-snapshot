import { cloneDeep, isFunction, isPlainObject } from 'lodash-es'
import { Settings } from './copypasta'
import { createAggregateError, ErrorLike, isErrorLike } from './errors'
import * as GQL from './schema/graphqlschema'
import { parseJSONCOrError } from './util'

export type ID = string

export interface IClient {
    __typename: 'Client'
    displayName: string
}

/**
 * A configuration subject is something that can have settings associated with it, such as a site ("global
 * settings"), an organization ("organization settings"), a user ("user settings"), etc.
 */
export type ConfigurationSubject = Pick<GQL.IConfigurationSubject, 'id' | 'settingsURL' | 'viewerCanAdminister'> &
    (
        | Pick<IClient, '__typename' | 'displayName'>
        | Pick<GQL.IUser, '__typename' | 'username' | 'displayName'>
        | Pick<GQL.IOrg, '__typename' | 'name' | 'displayName'>
        | Pick<GQL.ISite, '__typename'>)

/**
 * A cascade of settings from multiple subjects, from lowest precedence to highest precedence, and the final
 * settings, merged in order of precedence from the settings for each subject in the cascade.
 *
 * @template S the configuration subject type
 * @template C the settings type
 */
export interface ConfigurationCascade<S extends ConfigurationSubject, C = Settings> {
    /**
     * The settings for each subject in the cascade, from lowest to highest precedence.
     */
    subjects: ConfiguredSubject<S, C>[]

    /**
     * The final settings (merged in order of precedence from the settings for each subject in the cascade), an
     * error (if any occurred while retrieving, parsing, or merging the settings), or null if there are no settings
     * from any of the subjects.
     */
    merged: C | ErrorLike | null
}

/**
 * A subject and its settings.
 *
 * @template S the configuration subject type
 * @template C the settings type
 */
export interface ConfiguredSubject<S extends ConfigurationSubject, C = Settings> {
    /** The subject. */
    subject: S

    /**
     * The subject's settings (if any), an error (if any occurred while retrieving or parsing the settings), or
     * null if there are no settings.
     */
    settings: C | ErrorLike | null
}

/** A minimal subset of a GraphQL ConfigurationSubject type that includes only the single contents value. */
export interface SubjectConfigurationContents {
    latestSettings: {
        configuration: {
            contents: string
        }
    } | null
}

/** Converts a GraphQL ConfigurationCascade value to a value of this library's ConfigurationCascade type. */
export function gqlToCascade<S extends ConfigurationSubject, C = Settings>({
    subjects,
}: {
    subjects: (S & SubjectConfigurationContents)[]
}): ConfigurationCascade<S, C> {
    const cascade: ConfigurationCascade<S, C> = { subjects: [], merged: null }
    const allSettings: C[] = []
    const allSettingsErrors: ErrorLike[] = []
    for (const subject of subjects) {
        const settings = subject.latestSettings && parseJSONCOrError<C>(subject.latestSettings.configuration.contents)
        cascade.subjects.push({ subject, settings })

        if (isErrorLike(settings)) {
            allSettingsErrors.push(settings)
        } else if (settings !== null) {
            allSettings.push(settings)
        }
    }

    if (allSettingsErrors.length > 0) {
        cascade.merged = createAggregateError(allSettingsErrors)
    } else {
        cascade.merged = mergeSettings<C>(allSettings)
    }

    return cascade
}

/**
 * Deeply merges the settings without modifying any of the input values. The array is ordered from lowest to
 * highest precedence in the merge.
 *
 * TODO(sqs): In the future, this will pass a CustomMergeFunctions value to merge.
 */
export function mergeSettings<C extends Settings>(values: C[]): C | null {
    if (values.length === 0) {
        return null
    }
    const target = cloneDeep(values[0])
    for (const value of values.slice(1)) {
        merge(target, value)
    }
    return target
}

export interface CustomMergeFunctions {
    [key: string]: (base: any, add: any) => any | CustomMergeFunctions
}

/**
 * Deeply merges add into base (modifying base). The merged value for a key path can be customized by providing a
 * function at the same key path in custom.
 *
 * Most callers should use mergeSettings, which uses the set of CustomMergeFunctions that are required to properly
 * merge settings.
 */
export function merge(base: any, add: any, custom?: CustomMergeFunctions): void {
    for (const key of Object.keys(add)) {
        if (key in base) {
            const customEntry = custom && custom[key]
            if (customEntry && isFunction(customEntry)) {
                base[key] = customEntry(base[key], add[key])
            } else if (isPlainObject(base[key]) && isPlainObject(add[key])) {
                merge(base[key], add[key], customEntry)
            } else {
                base[key] = add[key]
            }
        } else {
            base[key] = add[key]
        }
    }
}

/**
 * The conventional ordering of extension configuration subject types in a list.
 */
export const SUBJECT_TYPE_ORDER: ConfigurationSubject['__typename'][] = ['Client', 'User', 'Org', 'Site']

export function subjectTypeHeader(nodeType: ConfigurationSubject['__typename']): string | null {
    switch (nodeType) {
        case 'Client':
            return null
        case 'Site':
            return null
        case 'Org':
            return 'Organization:'
        case 'User':
            return null
    }
}

export function subjectLabel(subject: ConfigurationSubject): string {
    switch (subject.__typename) {
        case 'Client':
            return 'Client'
        case 'Site':
            return 'Everyone'
        case 'Org':
            return subject.name
        case 'User':
            return subject.username
    }
}
