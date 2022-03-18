import {
    SettingsOrgSubject,
    SettingsSiteSubject,
    SettingsSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'

/**
 * Currently we support only two types of subject that can have insights dashboard.
 */
export type SupportedInsightSubject = SettingsUserSubject | SettingsOrgSubject | SettingsSiteSubject

/**
 * Supported insight subject types.
 *
 * Values of this enum are synced with settings subject __typename values.
 */
export enum SupportedInsightSubjectType {
    User = 'User',
    Organization = 'Org',
    Global = 'Site',
}

export const isGlobalSubject = (subject: SettingsSubject): subject is SettingsSiteSubject =>
    subject.__typename === SupportedInsightSubjectType.Global

export const isOrganizationSubject = (subject: SettingsSubject): subject is SettingsOrgSubject =>
    subject.__typename === SupportedInsightSubjectType.Organization

export const isUserSubject = (subject: SettingsSubject): subject is SettingsUserSubject =>
    subject.__typename === SupportedInsightSubjectType.User
