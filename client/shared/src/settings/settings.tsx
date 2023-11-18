import { createContext, useContext, useMemo } from 'react'

import { cloneDeep, isFunction } from 'lodash'

import { createAggregateError, type ErrorLike, isErrorLike, logger, parseJSONCOrError } from '@sourcegraph/common'

import type {
    DefaultSettingFields,
    OrgSettingFields,
    SiteSettingFields,
    UserSettingFields,
} from '../graphql-operations'
import type { Settings as GeneratedSettingsType, SettingsExperimentalFeatures } from '../schema/settings.schema'

/**
 * A dummy type to represent the "subject" for client settings (i.e., settings stored in the client application,
 * such as the browser extension). This subject doesn't exist in the GraphQL API, but the related types that are
 * also used as settings subjects {@link GQL.IUser}, {@link GQL.IOrg}, and {@link GQL.ISite} do.
 */
export interface IClient {
    __typename: 'Client'
    displayName: string
}

/**
 * A JSON Settings Schema type containing properties to prevent its misuse in a place of SettingsCascade.
 */
export interface Settings extends GeneratedSettingsType {
    [key: string]: any

    // These properties should never exist on Settings but do exist on SettingsCascade. This makes it so the
    // compiler points out where we misuse a Settings value in place of a SettingsCascade value and vice
    // versa.
    subjects?: never
    merged?: never // deprecated name, but keep it around
    final?: never
}

export type SettingsSubjectCommonFields = Pick<DefaultSettingFields, 'id' | 'viewerCanAdminister'>

export type ClientSettingFields = Pick<IClient, '__typename' | 'displayName'> &
    Pick<DefaultSettingFields, 'latestSettings'> &
    SettingsSubjectCommonFields

/**
 * A settings subject is something that can have settings associated with it, such as a site ("global
 * settings"), an organization ("organization settings"), a user ("user settings"), etc.
 */
export type SettingsSubject =
    | ClientSettingFields
    | UserSettingFields
    | OrgSettingFields
    | SiteSettingFields
    | DefaultSettingFields

/**
 * A cascade of settings from multiple subjects, from lowest precedence to highest precedence, and the final
 * settings, merged in order of precedence from the settings for each subject in the cascade.
 *
 * For example, the client might support settings globally and per-user, and it is designed so that
 * user settings override global settings. Then there would be two subjects, one for global settings and one for
 * the user.
 *
 * Callers that need to represent the null/error states should use {@link SettingsCascade}.
 *
 * @template S the settings type
 */
export interface SettingsCascade<S extends Settings = Settings> {
    /**
     * The settings for each subject in the cascade, from lowest to highest precedence.
     */
    subjects: ConfiguredSubject<S>[]

    /**
     * The final settings (merged in order of precedence from the settings for each subject in the cascade).
     */
    final: S
}

/**
 * A settings cascade that also supports representing subjects with no settings or whose settings triggered an
 * error.
 *
 * Callers that don't need to represent the null/error states should use {@link SettingsCascade}.
 *
 * @template S the settings type
 */
export interface SettingsCascadeOrError<S extends Settings = Settings> {
    /**
     * The settings for each subject in the cascade, from lowest to highest precedence, null if there are none, or
     * an error.
     *
     * @see SettingsCascade#subjects
     */
    subjects: ConfiguredSubjectOrError<S>[] | null

    /**
     * The final settings (merged in order of precedence from the settings for each subject in the cascade), an
     * error (if any occurred while retrieving, parsing, or merging the settings), or null if there are no settings
     * from any of the subjects.
     *
     * @see SettingsCascade#final
     */
    final: S | ErrorLike | null
}

export const EMPTY_SETTINGS_CASCADE: SettingsCascade = { final: {}, subjects: [] }

/**
 * A subject and its settings.
 *
 * Callers that need to represent the null/error states should use {@link ConfiguredSubjectOrError}.
 *
 * @template S the settings type
 */
interface ConfiguredSubject<S extends Settings = Settings> {
    /** The subject. */
    subject: SettingsSubject

    /** The subject's settings. */
    settings: S | null

    /** The sequential ID number of the settings, used to ensure that edits are applied to the correct version. */
    lastID: number | null
}

/**
 * A subject and its settings, or null if there are no settings, or an error.
 *
 * Callers that don't need to represent the null/error states should use {@link ConfiguredSubject}.
 *
 * @template S the settings type
 */
export interface ConfiguredSubjectOrError<S extends Settings = Settings>
    extends Pick<ConfiguredSubject<S>, Exclude<keyof ConfiguredSubject<S>, 'settings'>> {
    /**
     * The subject's settings (if any), an error (if any occurred while retrieving or parsing the settings), or
     * null if there are no settings.
     */
    settings: S | ErrorLike | null
}

/**
 * Converts a GraphQL SettingsCascade value to a SettingsCascadeOrError value.
 *
 * @param subjects A list of settings subjects in the settings cascade. If empty, an error is thrown.
 */
export function gqlToCascade({ subjects }: { subjects: SettingsSubject[] }): SettingsCascadeOrError {
    const configuredSubjects: ConfiguredSubjectOrError[] = []
    const allSettings: Settings[] = []
    const allSettingsErrors: ErrorLike[] = []
    for (const subject of subjects) {
        const settings = subject.latestSettings && parseJSONCOrError<Settings>(subject.latestSettings.contents)
        const lastID = subject.latestSettings ? subject.latestSettings.id : null
        configuredSubjects.push({ subject, settings, lastID })
        if (isErrorLike(settings)) {
            allSettingsErrors.push(settings)
        } else if (settings !== null) {
            allSettings.push(settings)
        }
    }
    return {
        subjects: configuredSubjects,
        final:
            allSettingsErrors.length > 0
                ? createAggregateError(allSettingsErrors)
                : mergeSettings<Settings>(allSettings),
    }
}

/**
 * Deeply merges the settings without modifying any of the input values. The array is ordered from lowest to
 * highest precedence in the merge.
 *
 * TODO(sqs): In the future, this will pass a CustomMergeFunctions value to merge.
 */
export function mergeSettings<S extends Settings>(values: S[]): S | null {
    if (values.length === 0) {
        return null
    }
    const customFunctions: CustomMergeFunctions = {
        extensions: (base: any, add: any) => ({ ...base, ...add }),
        experimentalFeatures: (base: any, add: any) => ({ ...base, ...add }),
        notices: (base: any, add: any) => [...base, ...add],
        'search.scopes': (base: any, add: any) => [...base, ...add],
        'search.savedQueries': (base: any, add: any) => [...base, ...add],
        quicklinks: (base: any, add: any) => [...base, ...add],
    }
    const target = cloneDeep(values[0])
    for (const value of values.slice(1)) {
        merge(target, value, customFunctions)
    }
    return target
}

export interface CustomMergeFunctions {
    [key: string]: (base: any, add: any) => any | CustomMergeFunctions
}

/**
 * Shallow merges add into base (modifying base). Only the top-level object is smerged.
 *
 * The merged value for a key path can be customized by providing a
 * function at the same key path in `custom`.
 *
 * Most callers should use mergeSettings, which uses the set of CustomMergeFunctions that are required to properly
 * merge settings.
 */
export function merge(base: any, add: any, custom?: CustomMergeFunctions): void {
    for (const key of Object.keys(add)) {
        if (key in base) {
            const customEntry = custom?.[key]
            if (customEntry && isFunction(customEntry)) {
                base[key] = customEntry(base[key], add[key])
            } else {
                base[key] = add[key]
            }
        } else {
            base[key] = add[key]
        }
    }
}

/**
 * Reports whether the settings cascade is valid (i.e., is non-empty and doesn't have any errors).
 *
 * @todo Display the errors to the user in another component.
 *
 * @template S the settings type
 * @deprecated Use useSettings() instead.
 */
export function isSettingsValid<S extends Settings>(
    settingsCascade: SettingsCascadeOrError<S>
): settingsCascade is SettingsCascade<S> {
    return (
        settingsCascade.subjects !== null &&
        !settingsCascade.subjects.some(subject => isErrorLike(subject.settings)) &&
        settingsCascade.final !== null &&
        !isErrorLike(settingsCascade.final)
    )
}

/**
 * React partial props for components needing the settings cascade.
 *
 * @deprecated Use useSettings() or useSettingsCascade() instead.
 */
export interface SettingsCascadeProps<S extends Settings = Settings> {
    settingsCascade: SettingsCascadeOrError<S>
}

interface SettingsContextData<S extends Settings = Settings> {
    settingsCascade: SettingsCascadeOrError<S>
}
const SettingsContext = createContext<SettingsContextData>({
    settingsCascade: EMPTY_SETTINGS_CASCADE,
})

interface SettingsProviderProps {
    settingsCascade: SettingsCascadeOrError
}

export const SettingsProvider: React.FC<React.PropsWithChildren<SettingsProviderProps>> = props => {
    const { children, settingsCascade } = props
    const context = useMemo(
        () => ({
            settingsCascade:
                // When the EMPTY_SETTINGS_CASCADE is used on purpose, we clone the object to avoid
                // it from mistakenly causing errors to be logged as if no context provider is used
                settingsCascade === EMPTY_SETTINGS_CASCADE ? { ...EMPTY_SETTINGS_CASCADE } : settingsCascade,
        }),
        [settingsCascade]
    )
    return <SettingsContext.Provider value={context}>{children}</SettingsContext.Provider>
}

/**
 * Access the underlying settings cascade directly.
 *
 * @deprecated Use useSettings() instead.
 */
export const useSettingsCascade = (): SettingsCascadeOrError => {
    const { settingsCascade } = useContext(SettingsContext)
    if (
        settingsCascade === EMPTY_SETTINGS_CASCADE &&
        (globalThis.process === undefined || process.env.VITEST_WORKER_ID === undefined)
    ) {
        logger.error(
            'useSettingsCascade must be used within a SettingsProvider, falling back to an empty settings object'
        )
    }
    return settingsCascade
}

/**
 * A React hooks that returns the resolved settings cascade.
 */
export const useSettings = (): Settings | null => {
    const settingsCascade = useSettingsCascade()
    return isSettingsValid(settingsCascade) ? settingsCascade.final : null
}

const defaultFeatures: SettingsExperimentalFeatures = {
    codeMonitoring: true,
    /**
     * Whether we show the multiline editor at /search/console
     */
    showMultilineSearchConsole: false,
    codeMonitoringWebHooks: true,
    showCodeMonitoringLogs: true,
    codeInsightsCompute: false,
    editor: 'codemirror6',
    codeInsightsRepoUI: 'search-query-or-strict-list',
    isInitialized: true,
    searchQueryInput: 'v2',
}

/**
 * A React hooks that can be used to query specific feature flags. Prioritize this over the generic
 * useSettings() hook if all you need is a feature flag.
 */
export function useExperimentalFeatures<T>(selector: (features: SettingsExperimentalFeatures) => T): T {
    const settings = useSettings()
    return selector({ ...defaultFeatures, ...settings?.experimentalFeatures })
}
