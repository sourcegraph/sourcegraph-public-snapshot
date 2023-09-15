import type { ApolloQueryResult } from '@apollo/client'
import type { SiteConfigResult } from 'src/graphql-operations'

import type { ErrorLike } from '@sourcegraph/common'

export interface OnboardingChecklistResult {
    licenseKey: LicenseInfo
    id: number
    config: string
    checklistItem: ChecklistItem
}

export interface LicenseInfo {
    key: string
    tags: string[]
    userCount: number
    expiresAt: string
}

export interface ChecklistItem {
    licenseKey: boolean
    externalURL: boolean
    emailSmtp: boolean
    authProviders: boolean
    externalServices: boolean
}

export interface UseOnboardingChecklistResult {
    loading: boolean
    error?: ErrorLike
    data?: OnboardingChecklistResult
    refetch?: () => Promise<ApolloQueryResult<SiteConfigResult>>
}

export interface EffectiveContent {
    licenseKey: string
    externalURL: string
    'email.smtp': {
        host: string
    }
    'auth.providers': string[]
}

export interface LicenseKeyInfo {
    title: string
    type: string
    description: string
    logo: JSX.Element
}
