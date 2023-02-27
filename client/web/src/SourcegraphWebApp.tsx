import 'focus-visible'

import { FC, useCallback, useEffect, useMemo, useState } from 'react'

import { ApolloProvider } from '@apollo/client'
import ServerIcon from 'mdi-react/ServerIcon'
import { RouterProvider, createBrowserRouter } from 'react-router-dom'
import { combineLatest, from, Subscription, fromEvent } from 'rxjs'

import { isTruthy, logger } from '@sourcegraph/common'
import { GraphQLClient, HTTPStatusError } from '@sourcegraph/http-client'
import { SharedSpanName, TraceSpanProvider } from '@sourcegraph/observability-client'
import { setCodeIntelSearchContext } from '@sourcegraph/shared/src/codeintel/searchContext'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { ShortcutProvider } from '@sourcegraph/shared/src/react-shortcuts'
import {
    isSearchContextSpecAvailable,
    SearchQueryStateStoreProvider,
    getDefaultSearchContextSpec,
} from '@sourcegraph/shared/src/search'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import {
    EMPTY_SETTINGS_CASCADE,
    Settings,
    SettingsCascadeOrError,
    SettingsProvider,
    SettingsSubjectCommonFields,
} from '@sourcegraph/shared/src/settings/settings'
import { TemporarySettingsProvider } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsProvider'
import { TemporarySettingsStorage } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsStorage'
import { FeedbackText, setLinkComponent, RouterLink, WildcardThemeContext, WildcardTheme } from '@sourcegraph/wildcard'

import { authenticatedUser as authenticatedUserSubject, AuthenticatedUser, authenticatedUserValue } from './auth'
import { getWebGraphQLClient } from './backend/graphql'
import { ComponentsComposer } from './components/ComponentsComposer'
import { ErrorBoundary } from './components/ErrorBoundary'
import { HeroPage } from './components/HeroPage'
import { FeatureFlagsProvider } from './featureFlags/FeatureFlagsProvider'
import { Layout } from './Layout'
import { LegacyRoute, LegacyRouteContextProvider } from './LegacyRouteContext'
import { createPlatformContext } from './platform/context'
import { parseSearchURL } from './search'
import { SearchResultsCacheProvider } from './search/results/SearchResultsCacheProvider'
import { GLOBAL_SEARCH_CONTEXT_SPEC } from './SearchQueryStateObserver'
import { StaticAppConfig } from './staticAppConfig'
import { setQueryStateFromSettings, setExperimentalFeaturesFromSettings, useNavbarQueryState } from './stores'
import { UserSessionStores } from './UserSessionStores'
import { siteSubjectNoAdmin, viewerSubjectFromSettings } from './util/settings'

import styles from './LegacySourcegraphWebApp.module.scss'

export interface StaticSourcegraphWebAppContext {
    setSelectedSearchContextSpec: (spec: string) => void
    platformContext: PlatformContext
    extensionsController: ExtensionsControllerProps['extensionsController'] | null
}

export interface DynamicSourcegraphWebAppContext {
    /**
     * TODO: Move all the search context logic as close as possible to the components
     * that actually need it. Remove related `useState` from the `SourcegraphWebApp` component.
     */
    selectedSearchContextSpec: string | undefined

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

    const [selectedSearchContextSpec, _setSelectedSearchContextSpec] = useState<string | undefined>()

    // NOTE(2022-09-08) Inform the inlined code from
    // sourcegraph/code-intel-extensions about the change of search context.
    // The old extension code previously accessed this information from the
    // 'sourcegraph' npm package, and updating the context like this was the
    // simplest solution to mirror the old behavior while deprecating
    // extensions on a tight deadline. It would be nice to properly pass
    // around this via React state in the future.
    const setWorkspaceSearchContext = useCallback((spec: string | null): void => {
        setCodeIntelSearchContext(spec ?? undefined)
    }, [])
    const setSelectedSearchContextSpecWithNoChecks = useCallback(
        (spec: string): void => {
            _setSelectedSearchContextSpec(spec)
            setWorkspaceSearchContext(spec)
        },
        [setWorkspaceSearchContext]
    )
    const setSelectedSearchContextSpecToDefault = useCallback((): void => {
        if (!props.searchContextsEnabled) {
            return
        }
        subscriptions.add(
            getDefaultSearchContextSpec({ platformContext }).subscribe(spec => {
                // Fall back to global if no default is returned.
                setSelectedSearchContextSpecWithNoChecks(spec || GLOBAL_SEARCH_CONTEXT_SPEC)
            })
        )
    }, [props.searchContextsEnabled, setSelectedSearchContextSpecWithNoChecks, subscriptions])

    const setSelectedSearchContextSpec = useCallback(
        (spec: string): void => {
            if (!props.searchContextsEnabled) {
                return
            }

            // The global search context is always available.
            if (spec === GLOBAL_SEARCH_CONTEXT_SPEC) {
                setSelectedSearchContextSpecWithNoChecks(spec)
            }

            // Check if the wanted search context is available.
            subscriptions.add(
                isSearchContextSpecAvailable({
                    spec,
                    platformContext,
                }).subscribe(isAvailable => {
                    if (isAvailable) {
                        setSelectedSearchContextSpecWithNoChecks(spec)
                    } else if (!selectedSearchContextSpec) {
                        // If the wanted search context is not available and
                        // there is no currently selected search context,
                        // set the current selection to the default search context.
                        // Otherwise, keep the current selection.
                        setSelectedSearchContextSpecToDefault()
                    }
                })
            )
        },
        [
            props.searchContextsEnabled,
            selectedSearchContextSpec,
            setSelectedSearchContextSpecToDefault,
            setSelectedSearchContextSpecWithNoChecks,
            subscriptions,
        ]
    )

    // TODO: Move all of this initialization outside React so we don't need to
    // handle the optional states everywhere
    useEffect(() => {
        const parsedSearchURL = parseSearchURL(window.location.search)
        const parsedSearchQuery = parsedSearchURL.query || ''

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
                    setExperimentalFeaturesFromSettings(settingsCascade)
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

        if (parsedSearchQuery && !filterExists(parsedSearchQuery, FilterType.context)) {
            // If a context filter does not exist in the query, we have to switch the selected context
            // to global to match the UI with the backend semantics (if no context is specified in the query,
            // the query is run in global context).
            setSelectedSearchContextSpecWithNoChecks(GLOBAL_SEARCH_CONTEXT_SPEC)
        }
        if (!parsedSearchQuery) {
            // If no query is present (e.g. search page, settings page),
            // select the user's default search context.
            setSelectedSearchContextSpecToDefault()
        }

        setWorkspaceSearchContext(selectedSearchContextSpec ?? null)

        return () => subscriptions.unsubscribe()

        // We only ever want to run this hook once when the component mounts for
        // parity with the old behavior.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    const staticContext = {
        setSelectedSearchContextSpec,
        platformContext,
        extensionsController: null,
    } satisfies StaticSourcegraphWebAppContext

    const dynamicContext = {
        selectedSearchContextSpec,
        authenticatedUser: resolvedAuthenticatedUser,
        viewerSubject,
        settingsCascade,
    } satisfies DynamicSourcegraphWebAppContext

    const router = useMemo(
        () =>
            createBrowserRouter([
                {
                    element: <LegacyRoute render={props => <Layout {...props} />} />,
                    children: props.routes.filter(isTruthy),
                },
            ]),
        [props.routes]
    )

    // TODO: move into a standalone component and reuse it between `SourcegraphWebApp` and `LegacySourcegraphWebApp`.
    if (window.pageError && window.pageError.statusCode !== 404) {
        const statusCode = window.pageError.statusCode
        const statusText = window.pageError.statusText
        const errorMessage = window.pageError.error
        const errorID = window.pageError.errorID

        let subtitle: JSX.Element | undefined
        if (errorID) {
            subtitle = <FeedbackText headerText="Sorry, there's been a problem." />
        }
        if (errorMessage) {
            subtitle = (
                <div className={styles.error}>
                    {subtitle}
                    {subtitle && <hr className="my-3" />}
                    <pre>{errorMessage}</pre>
                </div>
            )
        } else {
            subtitle = <div className={styles.error}>{subtitle}</div>
        }

        return <HeroPage icon={ServerIcon} title={`${statusCode}: ${statusText}`} subtitle={subtitle} />
    }

    if (graphqlClient === null || temporarySettingsStorage === null) {
        return null
    }

    return (
        <ComponentsComposer
            components={[
                // `ComponentsComposer` provides children via `React.cloneElement`.
                /* eslint-disable react/no-children-prop, react/jsx-key */
                <ApolloProvider client={graphqlClient} children={undefined} />,
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
                /* eslint-enable react/no-children-prop, react/jsx-key */
            ]}
        >
            <RouterProvider router={router} />
            <UserSessionStores />
        </ComponentsComposer>
    )
}
