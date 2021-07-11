import { SettingsOrgSubject, SettingsSubject, SettingsUserSubject } from '@sourcegraph/shared/src/settings/settings';

/**
 * Currently we support only two types of subject that can have insights dashboard.
 */
export type SupportedInsightSubject = SettingsUserSubject | SettingsOrgSubject

export enum SupportedInsightSubjectType {
    User = 'User',
    Organization = 'Org'
}

export const SUPPORTED_TYPES_OF_SUBJECT = new Set(Object.keys(SupportedInsightSubjectType))

export const isSubjectInsightSupported = (subject: SettingsSubject): subject is SupportedInsightSubject =>
    SUPPORTED_TYPES_OF_SUBJECT.has(subject.__typename)

export const isOrganizationSubject = (subject: SupportedInsightSubject): subject is SettingsOrgSubject =>
    subject.__typename === SupportedInsightSubjectType.Organization

export const isUserSubject = (subject: SupportedInsightSubject): subject is SettingsUserSubject =>
    subject.__typename === SupportedInsightSubjectType.User
