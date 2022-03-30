import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql, useQuery } from '@sourcegraph/http-client'

import { requestGraphQL } from '../backend/graphql'
import {
    FetchFeatureFlagsResult,
    OrgFeatureFlagOverridesResult,
    OrgFeatureFlagOverridesVariables,
} from '../graphql-operations'

import { getOverrideKey } from './lib/getOverrideKey'

class ProxyMap<K extends string, V extends boolean> extends Map<K, V> {
    constructor(private getter?: (key: K, value: V | undefined) => V | undefined) {
        super()
    }
    public get(key: K): V | undefined {
        const originalValue = super.get(key)
        return this.getter ? this.getter(key, originalValue) : originalValue
    }
}

// A union of all feature flags we currently have.
// If there are no feature flags at the moment, this should be `never`.
export type FeatureFlagName = 'open-beta-enabled' | 'quick-start-tour-for-authenticated-users'

export type FlagSet = ProxyMap<FeatureFlagName, boolean>

/**
 * Fetches the evaluated feature flags for the current user
 */
export function fetchFeatureFlags(): Observable<FlagSet> {
    return from(
        requestGraphQL<FetchFeatureFlagsResult>(
            gql`
                query FetchFeatureFlags {
                    viewerFeatureFlags {
                        name
                        value
                    }
                }
            `
        )
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            const result = new ProxyMap<FeatureFlagName, boolean>((key: FeatureFlagName, value?: boolean):
                | boolean
                | undefined => {
                const overriddenValue = localStorage.getItem(getOverrideKey(key))

                return overriddenValue !== 'undefined' &&
                    overriddenValue !== null &&
                    ['true', 'false'].includes(overriddenValue)
                    ? (JSON.parse(overriddenValue) as boolean)
                    : value
            })
            for (const flag of data.viewerFeatureFlags) {
                result.set(flag.name as FeatureFlagName, flag.value)
            }
            return result
        })
    )
}

export interface OrgFlagOverride {
    orgID: string
    flagName: string
    value: boolean
}

/**
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

export interface FeatureFlagProps {
    /**
     * Evaluated feature flags for the current viewer
     */
    featureFlags: FlagSet
}

export const EMPTY_FEATURE_FLAGS = new Map<FeatureFlagName, boolean>()
