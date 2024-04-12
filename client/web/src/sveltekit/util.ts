import { ApolloClient, gql } from '@apollo/client'
import { memoize } from 'lodash'

import { PageRoutes, CommunityPageRoutes } from '../routes.constants'

// List of all top level routes (excluding the index route and the repository catch-all route)
const allRoutes: string[] = (Object.values(PageRoutes) as string[]).concat(Object.values(CommunityPageRoutes)).filter(
    route =>
        // Remove the repository catch-all route because it matches everything and might match other routes
        // that are not actually the repo container route.
        route !== PageRoutes.RepoContainer &&
        // Remove index route because it will be a prefix of every pathname
        route !== PageRoutes.Index
)

/**
 * Returns true if SvelteKit is enabled for the given pathname.
 * In that case the caller should trigger a page reload to load the SvelteKit app.
 */
export const isEnabledRoute = memoize((pathname: string): boolean => {
    // Maps server route names to path regex patterns. These are the routes for which
    // the server will render the SvelteKit app.
    const enabledRoutes = window.context?.svelteKit?.enabledRoutes ?? []
    for (const route of enabledRoutes) {
        if (new RegExp(route).test(pathname)) {
            // TODO: handle repo root catch all
            return true
        }
    }
    return false
})

/**
 * Returns true if the SvelteKit app supports the given pathname, irrespective of whether it is enabled or not.
 */
export const isSupportedRoute = memoize((pathname: string): boolean => {
    // Maps server route names to path regex patterns. These are the routes for which
    // the server will render the SvelteKit app.
    const availableRoutes = new Set(window.context?.svelteKit?.availableRoutes ?? [])
    for (const route of availableRoutes) {
        if (new RegExp(route).test(pathname)) {
            // TODO: handle repo root catch all
            return true
        }
    }
    return false
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
