import { cloneDeep, isFunction, isPlainObject } from 'lodash-es'
import * as GQL from '../graphql/schema'
import { createAggregateError, ErrorLike, isErrorLike } from '../util/errors'
import { parseJSONCOrError } from '../util/jsonc'

export interface IClient {
    __typename: 'Client'
    displayName: string
}

/**
 * A subset of the settings JSON Schema type containing the minimum needed by this library.
 */
export interface Settings {
    extensions?: { [extensionID: string]: boolean }
    [key: string]: any

    // These properties should never exist on Settings but do exist on SettingsCascade. This makes it so the
    // compiler points out where we misuse a Settings value in place of a SettingsCascade value and vice
    // versa.
    subjects?: never
    merged?: never // deprecated name, but keep it around
    final?: never
}

/**
 * A settings subject is something that can have settings associated with it, such as a site ("global
 * settings"), an organization ("organization settings"), a user ("user settings"), etc.
 */
export type SettingsSubject = Pick<GQL.ISettingsSubject, 'id' | 'settingsURL' | 'viewerCanAdminister'> &
    (
        | Pick<IClient, '__typename' | 'displayName'>
        | Pick<GQL.IUser, '__typename' | 'username' | 'displayName'>
        | Pick<GQL.IOrg, '__typename' | 'name' | 'displayName'>
        | Pick<GQL.ISite, '__typename'>)

/**
 * A cascade of settings from multiple subjects, from lowest precedence to highest precedence, and the final
 * settings, merged in order of precedence from the settings for each subject in the cascade.
 *
 * Callers that need to represent the null/error states should use {@link SettingsCascade}.
 */
export interface SettingsCascade {
    /**
     * The settings for each subject in the cascade, from lowest to highest precedence.
     */
    subjects: ConfiguredSubject[]

    final: Settings
}

/**
 * A settings cascade that also supports representing subjects with no settings or whose settings triggered an
 * error.
 *
 * Callers that don't need to represent the null/error states should use {@link SettingsCascade}.
 */
export interface SettingsCascadeOrError
    extends Pick<SettingsCascade, Exclude<keyof SettingsCascade, 'subjects' | 'final'>> {
    /**
     * The settings for each subject in the cascade, from lowest to highest precedence, null if there are none, or
     * an error.
     *
     * @see SettingsCascade#subjects
     */
    subjects: ConfiguredSubjectOrError[] | ErrorLike | null

    /**
     * The final settings (merged in order of precedence from the settings for each subject in the cascade), an
     * error (if any occurred while retrieving, parsing, or merging the settings), or null if there are no settings
     * from any of the subjects.
     *
     * @see SettingsCascade#final
     */
    final: Settings | ErrorLike | null
}

/**
 * A subject and its settings.
 *
 * Callers that need to represent the null/error states should use {@link ConfiguredSubjectOrError}.
 */
export interface ConfiguredSubject {
    /** The subject. */
    subject: SettingsSubject

    /** The subject's settings. */
    settings: Settings

    /** The sequential ID number of the settings, used to ensure that edits are applied to the correct version. */
    lastID: number | null
}

/**
 * A subject and its settings, or null if there are no settings, or an error.
 *
 * Callers that don't need to represent the null/error states should use {@link ConfiguredSubject}.
 */
export interface ConfiguredSubjectOrError
    extends Pick<ConfiguredSubject, Exclude<keyof ConfiguredSubject, 'settings'>> {
    /**
     * The subject's settings (if any), an error (if any occurred while retrieving or parsing the settings), or
     * null if there are no settings.
     */
    settings: Settings | ErrorLike | null
}

/** A minimal subset of a GraphQL SettingsSubject type that includes only the single contents value. */
export interface SubjectSettingsContents {
    latestSettings: {
        id: number
        contents: string
    } | null
}

/** Converts a GraphQL SettingsCascade value to a value of this library's SettingsCascade type. */
export function gqlToCascade({
    subjects,
}: {
    subjects: (SettingsSubject & SubjectSettingsContents)[]
}): SettingsCascadeOrError {
    const cascade: SettingsCascadeOrError & { subjects: ConfiguredSubjectOrError[] } = {
        subjects: [],
        final: null,
    }
    const allSettings: Settings[] = []
    const allSettingsErrors: ErrorLike[] = []
    for (const subject of subjects) {
        const settings = subject.latestSettings && parseJSONCOrError<Settings>(subject.latestSettings.contents)
        const lastID = subject.latestSettings ? subject.latestSettings.id : null
        cascade.subjects.push({ subject, settings, lastID })

        if (isErrorLike(settings)) {
            allSettingsErrors.push(settings)
        } else if (settings !== null) {
            allSettings.push(settings)
        }
    }

    if (allSettingsErrors.length > 0) {
        cascade.final = createAggregateError(allSettingsErrors)
    } else {
        cascade.final = mergeSettings<Settings>(allSettings)
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
 * The conventional ordering of extension settings subject types in a list.
 */
export const SUBJECT_TYPE_ORDER: SettingsSubject['__typename'][] = ['Client', 'User', 'Org', 'Site']

export function subjectTypeHeader(nodeType: SettingsSubject['__typename']): string | null {
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

export function subjectLabel(subject: SettingsSubject): string {
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

/**
 * React partial props for components needing the settings cascade.
 */
export interface SettingsCascadeProps {
    settingsCascade: SettingsCascadeOrError
}
