import { gql, useQuery } from '@sourcegraph/http-client'

import type { OrgFeatureFlagOverridesResult, OrgFeatureFlagOverridesVariables } from '../graphql-operations'

// A union of all feature flags we currently have.
export const FEATURE_FLAGS = [
    'quick-start-tour-for-authenticated-users',
    'insight-polling-enabled',
    'ab-visitor-tour-with-notebooks',
    'ab-email-verification-alert',
    'contrast-compliant-syntax-highlighting',
    'admin-analytics-cache-disabled',
    'search-input-show-history',
    'search-results-keyboard-navigation',
    'enable-streaming-git-blame',
    'plg-enable-add-codehost-widget',
    'accessible-file-tree',
    'accessible-symbol-tree',
    'accessible-file-tree-always-load-ancestors',
    'enable-ownership-panels',
    'blob-page-switch-areas-shortcuts',
    'clone-progress-logging',
    'sourcegraph-operator-site-admin-hide-maintenance',
    'repository-metadata',
    'cody-web-search',
    'own-promote',
    'own-analytics',
    'enable-simple-search',
    'end-user-onboarding',
    'admin-onboarding',
    'enable-sveltekit',
    'search-content-based-lang-detection',
    'search-new-keyword',
    'search-debug',
    'cody-chat-mock-test',
    'signup-survey-enabled',
    'cody-pro',
] as const

export type FeatureFlagName = typeof FEATURE_FLAGS[number]

interface OrgFlagOverride {
    orgID: string
    flagName: string
    value: boolean
}

/**
 * // TODO: clarify why to use this if GQL already takes care of overrides?
 * Fetches all feature flag overrides for organizations that the current user is a member of
 */
export function useFlagsOverrides(): { data: OrgFlagOverride[]; loading: boolean } {
    const { data, loading } = useQuery<OrgFeatureFlagOverridesResult, OrgFeatureFlagOverridesVariables>(
        gql`
            query OrgFeatureFlagOverrides {
                organizationFeatureFlagOverrides {
                    namespace {
                        id
                    }
                    targetFlag {
                        ... on FeatureFlagBoolean {
                            name
                        }
                        ... on FeatureFlagRollout {
                            name
                        }
                    }
                    value
                }
            }
        `,
        { fetchPolicy: 'cache-and-network' }
    )

    if (!data) {
        return { data: [], loading }
    }

    return {
        data: data?.organizationFeatureFlagOverrides.map(value => ({
            orgID: value.namespace.id,
            flagName: value.targetFlag.name,
            value: value.value,
        })),
        loading,
    }
}
