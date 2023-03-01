import 'focus-visible'

import { FC, useEffect, useMemo, useState } from 'react'

import { ApolloProvider, SuspenseCache } from '@apollo/client'
import { RouterProvider, createBrowserRouter } from 'react-router-dom'
import { combineLatest, from, Subscription, fromEvent } from 'rxjs'

import { logger } from '@sourcegraph/common'
import { GraphQLClient, HTTPStatusError } from '@sourcegraph/http-client'
import { SharedSpanName, TraceSpanProvider } from '@sourcegraph/observability-client'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { ShortcutProvider } from '@sourcegraph/shared/src/react-shortcuts'
import { SearchQueryStateStoreProvider } from '@sourcegraph/shared/src/search'
import {
    EMPTY_SETTINGS_CASCADE,
    Settings,
    SettingsCascadeOrError,
    SettingsProvider,
    SettingsSubjectCommonFields,
} from '@sourcegraph/shared/src/settings/settings'
import { TemporarySettingsProvider } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsProvider'
import { TemporarySettingsStorage } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsStorage'
import { setLinkComponent, RouterLink, WildcardThemeContext, WildcardTheme } from '@sourcegraph/wildcard'

import { authenticatedUser as authenticatedUserSubject, AuthenticatedUser, authenticatedUserValue } from './auth'
import { getWebGraphQLClient } from './backend/graphql'
import { ComponentsComposer } from './components/ComponentsComposer'
import { ErrorBoundary, RouteError } from './components/ErrorBoundary'
import { FeatureFlagsProvider } from './featureFlags/FeatureFlagsProvider'
import { Layout } from './Layout'
import { LegacyRoute, LegacyRouteContextProvider } from './LegacyRouteContext'
import { PageError } from './PageError'
import { createPlatformContext } from './platform/context'
import { SearchResultsCacheProvider } from './search/results/SearchResultsCacheProvider'
import { StaticAppConfig } from './staticAppConfig'
import { setQueryStateFromSettings, useNavbarQueryState } from './stores'
import { UserSessionStores } from './UserSessionStores'
import { siteSubjectNoAdmin, viewerSubjectFromSettings } from './util/settings'
import { SearchContextProvider } from './SearchContext'

export interface StaticSourcegraphWebAppContext {
    platformContext: PlatformContext
    extensionsController: ExtensionsControllerProps['extensionsController'] | null
}

export interface DynamicSourcegraphWebAppContext {
    /**
     * TODO:
     * 1. Move `authenticatedUser` to Apollo Client.
     * 2. Remove `resolvedAuthenticatedUser` from the `SourcegraphWebApp` component
     * 3. Initialize `authenticatedUser` in the Apollo Client cache on application load from `window.context.currentUser`.
     * 4. Remove it from prop drilling and use the `useQuery` hook to get from it the Apollo client context.
     */
    authenticatedUser: AuthenticatedUser | null

    /**
     * TODO: Move `settingsCascade` out of rxjs. Probably, we can still keep rxjs wrapper
     * in the `platformContext` to avoid huge refactorings in non-Storm components
     * but the flow in `SourcegraphWebApp` needs to rely on the Apollo Client to untangle
     * subscriptions logic.
     *
     * Note: we already use Apollo Client to fetch settings inside of rxjs.
     */
    settingsCascade: SettingsCascadeOrError<Settings>

    /**
     * Computed from `settingsCascade` and `authenticatedUser`.
     */
    viewerSubject: SettingsSubjectCommonFields
}

const WILDCARD_THEME: WildcardTheme = {
    isBranded: true,
}

setLinkComponent(RouterLink)

const suspenseCache = new SuspenseCache()

/**
 * The synchronous and static value that creates the `platformContext.settings`
 * observable that sends the API request to the server to get `viewerSettings`.
 *
 * Most of the dynamic values in the `SourcegraphWebApp` depend on this observable.
 */
const platformContext = createPlatformContext()

export const SourcegraphWebApp: FC<StaticAppConfig> = props => {
    const [subscriptions] = useState(() => new Subscription())

    const [resolvedAuthenticatedUser, setResolvedAuthenticatedUser] = useState<AuthenticatedUser | null>(
        authenticatedUserValue
    )

    /**
     * TODO: Remove this state and get this data from the Apollo Client cache.
     * It's already available there because we rely on `client.watchQuery` in `createPlatformContext`.
     */
    const [settingsCascade, setSettingsCascade] = useState<SettingsCascadeOrError<Settings>>(EMPTY_SETTINGS_CASCADE)
    const [viewerSubject, setViewerSubject] = useState<SettingsSubjectCommonFields>(() => siteSubjectNoAdmin())

    /**
     * TODO: Make it synchrounously available in the `SourcegraphWebApp` component to remove redundant `useState`s
     * for the `graphqlClient` and `temporarySettingsStorage`.
     */
    const [graphqlClient, setGraphqlClient] = useState<GraphQLClient | null>(null)
    const [temporarySettingsStorage, setTemporarySettingsStorage] = useState<TemporarySettingsStorage | null>(null)

    // TODO: Move all of this initialization outside React so we don't need to
    // handle the optional states everywhere
    useEffect(() => {
        getWebGraphQLClient()
            .then(graphqlClient => {
                setGraphqlClient(graphqlClient)
                setTemporarySettingsStorage(
                    new TemporarySettingsStorage(graphqlClient, window.context.isAuthenticatedUser)
                )
            })
            .catch(error => {
                logger.error('Error initializing GraphQL client', error)
            })

        subscriptions.add(
            combineLatest([from(platformContext.settings), authenticatedUserSubject]).subscribe(
                ([settingsCascade, authenticatedUser]) => {
                    setQueryStateFromSettings(settingsCascade)
                    setSettingsCascade(settingsCascade)
                    setResolvedAuthenticatedUser(authenticatedUser ?? null)
                    setViewerSubject(viewerSubjectFromSettings(settingsCascade, authenticatedUser))
                }
            )
        )

        /**
         * TODO: move outiside of React and remove redundant rxjs wrapper.
         *
         * Listens for uncaught 401 errors when a user when a user was previously authenticated.
         *
         * Don't subscribe to this event when there wasn't an authenticated user,
         * as it could lead to an infinite loop of 401 -> reload -> 401
         */
        if (window.context.isAuthenticatedUser) {
            subscriptions.add(
                fromEvent<ErrorEvent>(window, 'error').subscribe(event => {
                    if (event?.error instanceof HTTPStatusError && event.error.status === 401) {
                        location.reload()
                    }
                })
            )
        }

        return () => subscriptions.unsubscribe()

        // We only ever want to run this hook once when the component mounts for
        // parity with the old behavior.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    const staticContext = {
        platformContext,
        extensionsController: null,
    } satisfies StaticSourcegraphWebAppContext

    const dynamicContext = {
        authenticatedUser: resolvedAuthenticatedUser,
        viewerSubject,
        settingsCascade,
    } satisfies DynamicSourcegraphWebAppContext

    const router = useMemo(
        () =>
            createBrowserRouter([
                {
                    element: <LegacyRoute render={props => <Layout {...props} />} />,
                    children: props.routes,
                    errorElement: <RouteError />,
                },
            ]),
        [props.routes]
    )

    const pageError = window.pageError
    if (pageError && pageError.statusCode !== 404) {
        return <PageError pageError={pageError} />
    }

    if (graphqlClient === null || temporarySettingsStorage === null) {
        return null
    }

    return (
        <ComponentsComposer
            components={[
                // `ComponentsComposer` provides children via `React.cloneElement`.
                /* eslint-disable react/no-children-prop, react/jsx-key */
                <ApolloProvider client={graphqlClient} children={undefined} suspenseCache={suspenseCache} />,
                <WildcardThemeContext.Provider value={WILDCARD_THEME} />,
                <SettingsProvider settingsCascade={settingsCascade} />,
                <ErrorBoundary location={null} />,
                <TraceSpanProvider name={SharedSpanName.AppMount} />,
                <FeatureFlagsProvider />,
                <ShortcutProvider />,
                <TemporarySettingsProvider temporarySettingsStorage={temporarySettingsStorage} />,
                <SearchResultsCacheProvider />,
                <SearchQueryStateStoreProvider useSearchQueryState={useNavbarQueryState} />,
                <LegacyRouteContextProvider
                    context={{
                        ...staticContext,
                        ...dynamicContext,
                        ...props,
                    }}
                />,
                <SearchContextProvider searchContextsEnabled={props.searchContextsEnabled} />,
                /* eslint-enable react/no-children-prop, react/jsx-key */
            ]}
        >
            <RouterProvider router={router} />
            <UserSessionStores />
        </ComponentsComposer>
    )
}
