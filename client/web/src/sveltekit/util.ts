import { ApolloClient, gql } from '@apollo/client'
import { memoize } from 'lodash'
import { matchPath } from 'react-router-dom'

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

// Mapping of route names to server side route names. Due to how client side routing works in this app not every
// server side route name has a 1:1 mapping to a client side route name. That's especially true for subroutes like
// the repository container route.
// ADD NEW SVELTEKIT SUPPORTED ROUTES HERE
const routeMap: { route: string; serverRoute: string; pathMatch?: RegExp }[] = [
    {
        route: PageRoutes.Search,
        serverRoute: 'search',
    },
    {
        route: PageRoutes.RepoContainer,
        serverRoute: 'repo',
        pathMatch: /^\/(.(?!\/-\/))+\/?$/,
    },
    {
        route: PageRoutes.RepoContainer,
        serverRoute: 'tree',
        pathMatch: /^\/.+?\/-\/tree\//,
    },
    {
        route: PageRoutes.RepoContainer,
        serverRoute: 'blob',
        pathMatch: /^\/.+?\/-\/blob\//,
    },
    {
        route: PageRoutes.RepoContainer,
        serverRoute: 'repo-commit',
        pathMatch: /^\/.+?\/-\/commit\//,
    },
    {
        route: PageRoutes.RepoContainer,
        serverRoute: 'repo-branches',
        pathMatch: /^\/.+?\/-\/branches\//,
    },
    {
        route: PageRoutes.RepoContainer,
        serverRoute: 'repo-commits',
        pathMatch: /^\/.+?\/-\/commits$/,
    },
    {
        route: PageRoutes.RepoContainer,
        serverRoute: 'repo-tags',
        pathMatch: /^\/.+?\/-\/tags$/,
    },
    {
        route: PageRoutes.RepoContainer,
        serverRoute: 'repo-stats',
        pathMatch: /^\/.+?\/-\/stats$/,
    },
]

function mapPathToServerRoute(pathname: string, serverRoutes: Set<string>): string | null {
    if (!pathname || serverRoutes.size === 0) {
        return null
    }

    // Determine whether the provided path is matching any of the known react-router routes.
    const targetRoute = allRoutes.find(
        route =>
            // Some routes in PageRoutes are not actually the exact paths passed to react router. Some are "extended"
            // in routes.tsx. For example, PageRoutes.CodyChat is used as PageRoutes.CodyChat + '/*' in routes.tsx.
            // But we cannot use routes.tsx directly here because it causes import ordering issues, specifically
            // for CSS.
            pathname.startsWith(route) || matchPath(route, pathname)
    )

    if (targetRoute) {
        // If we found a matching route we need to translate it to a server route name and check whether it's in the list of
        // enabled routes.
        for (const { route, serverRoute, pathMatch } of routeMap) {
            if (targetRoute === route && (!pathMatch || pathMatch.test(pathname)) && serverRoutes.has(serverRoute)) {
                return serverRoute
            }
        }
    } else {
        // If we didn't find a matching route we have to assume the path is a repository container route because
        // that is the catch-all route
        for (const { route, serverRoute, pathMatch } of routeMap) {
            if (
                PageRoutes.RepoContainer === route &&
                (!pathMatch || pathMatch.test(pathname)) &&
                serverRoutes.has(serverRoute)
            ) {
                return serverRoute
            }
        }
    }
    return null
}

/**
 * Returns true if SvelteKit is enabled for the given pathname.
 * In that case the caller should trigger a page reload to load the SvelteKit app.
 */
export const isEnabledRoute = memoize((pathname: string): boolean => {
    // Maps server route names to path regex patterns. These are the routes for which
    // the server will render the SvelteKit app.
    const enabledRoutes = new Set(window.context?.svelteKit?.enabledRoutes ?? [])
    return mapPathToServerRoute(pathname, enabledRoutes) !== null
})

/**
 * Returns true if the SvelteKit app supports the given pathname, irrespective of whether it is enabled or not.
 */
export const isSupportedRoute = memoize((pathname: string): boolean => {
    // Maps server route names to path regex patterns. These are the routes for which
    // the server will render the SvelteKit app.
    const availableRoutes = new Set(window.context?.svelteKit?.availableRoutes ?? [])
    return mapPathToServerRoute(pathname, availableRoutes) !== null
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
