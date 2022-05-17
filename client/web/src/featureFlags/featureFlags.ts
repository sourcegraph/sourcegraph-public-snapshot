import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql, useQuery } from '@sourcegraph/http-client'

import { requestGraphQL } from '../backend/graphql'
import {
    OrgFeatureFlagOverridesResult,
    OrgFeatureFlagOverridesVariables,
    EvaluatedFeatureFlagsResult,
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

const featureFlagCache = new ProxyMap<FeatureFlagName, boolean>((key: FeatureFlagName, value?: boolean):
    | boolean
    | undefined => {
    const overriddenValue = localStorage.getItem(getOverrideKey(key))

    return overriddenValue !== 'undefined' && overriddenValue !== null && ['true', 'false'].includes(overriddenValue)
        ? (JSON.parse(overriddenValue) as boolean)
        : value
})

// A union of all feature flags we currently have.
// If there are no feature flags at the moment, this should be `never`.
export type FeatureFlagName = 'open-beta-enabled' | 'quick-start-tour-for-authenticated-users' | 'new-repo-page'

type FlagSet = ProxyMap<FeatureFlagName, boolean>

/**
 * Fetches the evaluated feature flags for the current user
 */
function fetchFeatureFlags(): Observable<FlagSet> {
    return from(
        requestGraphQL<EvaluatedFeatureFlagsResult>(
            gql`
                query EvaluatedFeatureFlags {
                    evaluatedFeatureFlags {
                        name
                        value
                    }
                }
            `
        )
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            for (const flag of data.evaluatedFeatureFlags) {
                featureFlagCache.set(flag.name as FeatureFlagName, flag.value)
            }
            return featureFlagCache
        })
    )
}

/**
 * Fetches the evaluated feature flags for the current user
 */
function fetchFeatureFlag(flagName: FeatureFlagName): Observable<boolean> {
    return from(
        requestGraphQL<boolean>(
            gql`
                query FetchFeatureFlag($flagName: String!) {
                    evaluateFeatureFlag(flagName: $flagName)
                }
            `,
            { flagName }
        )
    ).pipe(map(dataOrThrowErrors))
}

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

export class FeatureFlagClient {
    private cache = new ProxyMap<FeatureFlagName, boolean>()

    constructor() {
            fetchFeatureFlags()
                .toPromise()
                .then(flags => (this.cache = flags))
                .catch(console.error)
    }

    /**
     * Evaluate a given feature flag for the current user or get from cache if already evaluated.
     */
    public async variation(flagName: FeatureFlagName): Promise<boolean> {
        if (!this.cache.has(flagName)) {
            const value = await fetchFeatureFlag(flagName).toPromise()
            this.cache.set(flagName, value)
        }

        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        return this.cache.get(flagName)!
    }

    /**
     * Calls callback function whenever a feature flag is changed
     *
     * @returns a cleanup/unsubscribe function
     */
    // eslint-disable-next-line id-length
    public on(flagName: FeatureFlagName, callback: (value: boolean, error?: Error) => void): () => void {
        this.variation(flagName)
            .then(callback)
            .catch(error => callback(false, error))
        return () => {
            // TODO: cleanup
        }
    }
}
