import { ApolloClient, gql } from '@apollo/client'
import { memoize } from 'lodash'

import { svelteKitRoutes } from './routes'

let knownRoutesRegex: RegExp | undefined

function getKnownRoutesRegex(): RegExp {
    if (!knownRoutesRegex) {
        knownRoutesRegex = new RegExp(`(${window.context?.svelteKit?.knownRoutes?.join(')|(')})`)
    }
    return knownRoutesRegex
}

function findSupportedRouteIndex(pathname: string): number {
    let index = -1

    for (let i = 0; i < svelteKitRoutes.length; i++) {
        const route = svelteKitRoutes[i]
        if (route.pattern.test(pathname)) {
            index = i
            if (!route.isRepoRoot) {
                break
            }
            // If the found route is the repo root we have to keep going
            // to find a more specific route.
        }
    }

    if (index !== -1) {
        // Check known routes to see if there is a more specific route
        // if yes then we should load the React app.
        // (if the more specific route is enabled)
        if (svelteKitRoutes[index].isRepoRoot && getKnownRoutesRegex().test(pathname)) {
            return -1
        }
    }

    return index
}

/**
 * Returns true if SvelteKit is enabled for the given pathname.
 * In that case the caller should trigger a page reload to load the SvelteKit app.
 */
export const isEnabledRoute = memoize((pathname: string): boolean => {
    // Maps server route names to path regex patterns. These are the routes for which
    // the server will render the SvelteKit app.
    const enabledRoutes = window.context?.svelteKit?.enabledRoutes ?? []
    const index = findSupportedRouteIndex(pathname)
    return index !== -1 && enabledRoutes.includes(index)
})

/**
 * Returns true if the SvelteKit app supports the given pathname, irrespective of whether it is enabled or not.
 */
export const canEnableSvelteKit = memoize((pathname: string): boolean => {
    if (!window.context?.svelteKit?.showToggle) {
        return false
    }
    return findSupportedRouteIndex(pathname) !== -1
})

export async function enableSvelteAndReload(client: ApolloClient<{}>, userID: string): Promise<void> {
    await client.mutate({
        mutation: gql`
            mutation EnableSveltePrototype($userID: ID!) {
                overrideWebNext: createFeatureFlagOverride(namespace: $userID, flagName: "web-next", value: true) {
                    __typename
                }
                overrideRollout: createFeatureFlagOverride(
                    namespace: $userID
                    flagName: "web-next-rollout"
                    value: true
                ) {
                    __typename
                }
            }
        `,
        variables: { userID },
    })
    window.location.reload()
}
