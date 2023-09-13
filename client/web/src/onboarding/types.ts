import type { Dispatch, SetStateAction } from 'react'

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
    usersPermissions: boolean
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

export interface OnboardingChecklistItemProps {
    isComplete: boolean
    title: string
    description: string
    link: string
}

export interface LicenseKeyModalProps {
    id: number
    config: string
    licenseKey: LicenseInfo
    refetch?: () => Promise<ApolloQueryResult<SiteConfigResult>>
    onHandleLicenseCheck: (
        newValue: boolean | ((previousValue: boolean | undefined) => boolean | undefined) | undefined
    ) => void
}

export interface LicenseKeyProps {
    isValid: boolean
    licenseInfo: LicenseInfo
}

export interface LicenseKeyInfo {
    title: string
    type: string
    description: string
    logo: JSX.Element
}

export interface OnboardingChecklistProps {
    isModalOpen: boolean
    onHandleOpen: Dispatch<SetStateAction<boolean>>
    keepOpen: boolean
    onHandleKeepOpen: Dispatch<SetStateAction<boolean>>
}
