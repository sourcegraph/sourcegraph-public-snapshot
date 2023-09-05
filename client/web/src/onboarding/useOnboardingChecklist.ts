import { parse } from 'jsonc-parser'
import type { SiteConfigResult, SiteConfigVariables } from 'src/graphql-operations'

import type { ErrorLike } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'

import { SITE_CONFIG_QUERY } from './queries'

interface OnboardingChecklistResult {
    licenseKey: boolean
    externalURL: boolean
    emailSmtp: boolean
    authProviders: boolean
    externalServices: boolean
}

interface UseOnboardingChecklistResult {
    loading: boolean
    error?: ErrorLike
    data?: OnboardingChecklistResult | undefined
}

export const useOnboardingChecklist = (): UseOnboardingChecklistResult => {
    const { loading, error, data } = useQuery<SiteConfigResult, SiteConfigVariables>(SITE_CONFIG_QUERY, {
        fetchPolicy: 'no-cache',
    })

    return {
        loading,
        error,
        data: data ? getChecklistItems(data) : undefined,
    }
}

interface EffectiveContent {
    licenseKey: string
    externalURL: string
    'email.smtp': {
        host: string
    }
    'auth.providers': string[]
}

function getChecklistItems(data: SiteConfigResult): OnboardingChecklistResult {
    const config = parse(data.site.configuration.effectiveContents) as EffectiveContent
    return {
        licenseKey: config.licenseKey !== '',
        externalURL: config.externalURL !== '',
        emailSmtp: config['email.smtp'].host !== '',
        authProviders: config['auth.providers'].length > 0,
        externalServices: data.externalServices?.nodes?.length > 0 || false,
    }
}
