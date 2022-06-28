import React, { useEffect, useState, useCallback } from 'react'

import * as H from 'history'
import { concat, of, timer } from 'rxjs'
import { debounce, delay, map, switchMap, takeUntil, tap, distinctUntilChanged } from 'rxjs/operators'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { createAggregateError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { ConfiguredRegistryExtension, isExtensionEnabled } from '@sourcegraph/shared/src/extensions/extension'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { ExtensionCategory, EXTENSION_CATEGORIES } from '@sourcegraph/shared/src/schema/extensionSchema'
import { Settings, SettingsCascadeProps, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { buildGetStartedURL } from '@sourcegraph/shared/src/util/url'
import { AlertLink, useLocalStorage, useEventObservable, Alert, Link, Input } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'
import {
    RegistryExtensionsResult,
    RegistryExtensionFieldsForList,
    RegistryExtensionsVariables,
} from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'

import { ExtensionBanner } from './ExtensionBanner'
import { ExtensionRegistrySidenav } from './ExtensionRegistrySidenav'
import {
    configureExtensionRegistry,
    ConfiguredExtensionRegistry,
    MinimalConfiguredRegistryExtension,
    configureFeaturedExtensions,
} from './extensions'
import { ExtensionsAreaRouteContext } from './ExtensionsArea'
import { ExtensionsList } from './ExtensionsList'

interface Props
    extends Pick<ExtensionsAreaRouteContext, 'authenticatedUser' | 'subject' | 'isSourcegraphDotCom'>,
        PlatformContextProps<'settings' | 'updateSettings' | 'requestGraphQL'>,
        SettingsCascadeProps,
        ThemeProps {
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
}

const LOADING = 'loading' as const
const URL_QUERY_PARAM = 'query'
const URL_CATEGORY_PARAM = 'category'
const SHOW_EXPERIMENTAL_EXTENSIONS_KEY = 'show-experimental-extensions'

export type ExtensionListData =
    | typeof LOADING
    | (ConfiguredExtensionRegistry & {
          featuredExtensions?: MinimalConfiguredRegistryExtension[]
          error: string | null
      })
    | ErrorLike

export type ExtensionsEnablement = 'all' | 'enabled' | 'disabled'

export type ExtensionCategoryOrAll = ExtensionCategory | 'All'

const extensionRegistryQuery = gql`
    query RegistryExtensions($query: String, $prioritizeExtensionIDs: [String!]!, $getFeatured: Boolean!) {
        extensionRegistry {
            extensions(query: $query, prioritizeExtensionIDs: $prioritizeExtensionIDs) {
                nodes {
                    ...RegistryExtensionFieldsForList
                }
                error
            }
            featuredExtensions @include(if: $getFeatured) {
                nodes {
                    ...RegistryExtensionFieldsForList
                }
                error
            }
        }
    }
    fragment RegistryExtensionFieldsForList on RegistryExtension {
        id
        publisher {
            __typename
            ... on User {
                id
                username
                displayName
                url
            }
            ... on Org {
                id
                name
                displayName
                url
            }
        }
        extensionID
        extensionIDWithoutRegistry
        name
        manifest {
            raw
            description
        }
        createdAt
        updatedAt
        url
        remoteURL
        registryName
        isLocal
        isWorkInProgress
        viewerCanAdminister
    }
`

export type ConfiguredExtensionCache = Map<string, MinimalConfiguredRegistryExtension>

/** A page that displays overview information about the available extensions. */
export const ExtensionRegistry: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    useEffect(() => eventLogger.logViewEvent('ExtensionsOverview'), [])

    const { history, location, settingsCascade, platformContext, authenticatedUser } = props

    // Update the cache after each response. This speeds up client-side filtering.
    // Lazy initialize cache ref. Don't mind the `useState` abuse:
    // - can't use `useMemo` here (https://github.com/facebook/react/issues/14490#issuecomment-454973512)
    // - `useRef` is just `useState`
    const [configuredExtensionCache] = useState<ConfiguredExtensionCache>(
        () => new Map<string, ConfiguredRegistryExtension<RegistryExtensionFieldsForList>>()
    )

    const [query, setQuery] = useState(getQueryFromLocation(location))

    const [selectedCategory, setSelectedCategory] = useState<ExtensionCategoryOrAll>(
        getCategoryFromLocation(location) || 'All'
    )

    // Filter extensions by enablement state: enabled, disabled, or all.
    const [enablementFilter, setEnablementFilter] = useState<ExtensionsEnablement>('all')

    // Used to determine `isLoading` in order to stop <ExtensionList> from filtering based on the
    // selected category before the new query has completed.
    // Note: It'll be worth refactoring the query pipeline (probably split between category changes and
    // query changes) when any more complexity is introduced.
    const [changedCategory, setChangedCategory] = useState(false)

    // Whether to show extensions that set "wip" to true in their manifests.
    const [storedShowExperimentalExtensions, setShowExperimentalExtensions] = useLocalStorage(
        SHOW_EXPERIMENTAL_EXTENSIONS_KEY,
        false
    )
    // Referrers can override user's "show experimental" setting by setting URL parameter 'experimental=true'
    const parameterShowExperimental = getExperimentalFromLocation(location)
    const showExperimentalExtensions = parameterShowExperimental || storedShowExperimentalExtensions
    const toggleExperimentalExtensions = (): void => {
        if (parameterShowExperimental) {
            // Clear search parameter when user clicks "show experimental" checkbox
            history.replace(
                getRegistryLocationDescriptor({
                    query: getQueryFromLocation(location),
                    category: getCategoryFromLocation(location),
                    experimental: false,
                })
            )
        }
        setShowExperimentalExtensions(!showExperimentalExtensions)
    }

    /**
     * Note: pass `settingsCascade` instead of making it a dependency to prevent creating
     * new subscriptions when user toggles extensions
     */
    const [nextQueryInput, data] = useEventObservable<
        {
            query: string
            category: ExtensionCategoryOrAll
            immediate: boolean
            settingsCascade: SettingsCascadeOrError<Settings>
        },
        ExtensionListData
    >(
        useCallback(
            newQueries =>
                newQueries.pipe(
                    distinctUntilChanged(
                        (previous, current) =>
                            previous.query === current.query && previous.category === current.category
                    ),
                    tap(({ query, category }) => {
                        setQuery(query)
                        setSelectedCategory(getCategoryFromLocation(window.location))

                        history.replace(
                            getRegistryLocationDescriptor({
                                query,
                                category,
                                experimental: getExperimentalFromLocation(window.location),
                            })
                        )
                    }),
                    debounce(({ immediate }) => timer(immediate ? 0 : 50)),
                    distinctUntilChanged(
                        (previous, current) =>
                            previous.query === current.query && previous.category === current.category
                    ),
                    switchMap(({ query, category, immediate, settingsCascade }) => {
                        let enabledExtensions: string[] = []
                        if (!isErrorLike(settingsCascade.final)) {
                            if (settingsCascade.final?.extensions) {
                                enabledExtensions = Object.keys(settingsCascade.final.extensions).filter(extensionID =>
                                    isExtensionEnabled(settingsCascade.final, extensionID)
                                )
                            }
                        }

                        if (category !== 'All') {
                            query = `${query} category:"${category}"`
                        }

                        // Only fetch + show featured extensions when there's no query or category selected.
                        const shouldGetFeaturedExtensions = category === 'All' && query.trim() === ''

                        const resultOrError = platformContext.requestGraphQL<
                            RegistryExtensionsResult,
                            RegistryExtensionsVariables
                        >({
                            request: extensionRegistryQuery,
                            variables: {
                                query,
                                prioritizeExtensionIDs: enabledExtensions,
                                getFeatured: shouldGetFeaturedExtensions,
                            },
                            mightContainPrivateInfo: true,
                        })

                        return concat(
                            of(LOADING).pipe(delay(immediate ? 0 : 250), takeUntil(resultOrError)),
                            resultOrError
                        )
                    }),
                    map(resultOrErrorOrLoading => {
                        if (resultOrErrorOrLoading === LOADING) {
                            return resultOrErrorOrLoading
                        }

                        const { data, errors } = resultOrErrorOrLoading

                        if (!data?.extensionRegistry?.extensions) {
                            return createAggregateError(errors)
                        }

                        const { error, nodes } = data.extensionRegistry.extensions

                        const featuredExtensions = data.extensionRegistry.featuredExtensions?.nodes
                            ? configureFeaturedExtensions(
                                  data.extensionRegistry.featuredExtensions.nodes,
                                  configuredExtensionCache
                              )
                            : undefined

                        return {
                            error,
                            featuredExtensions,
                            ...configureExtensionRegistry(nodes, configuredExtensionCache),
                        }
                    }),
                    tap(() => {
                        // In case this query was triggered due to a category change
                        setChangedCategory(false)
                    })
                ),
            [platformContext, history, configuredExtensionCache]
        )
    )

    const onQueryChangeEvent = useCallback(
        (event: React.FormEvent<HTMLInputElement>) =>
            nextQueryInput({
                query: event.currentTarget.value,
                category: getCategoryFromLocation(window.location),
                immediate: false,
                settingsCascade,
            }),
        [nextQueryInput, settingsCascade]
    )

    const onQueryChangeImmediate = useCallback(
        () =>
            nextQueryInput({
                query: getQueryFromLocation(window.location),
                category: getCategoryFromLocation(window.location),
                immediate: true,
                settingsCascade,
            }),
        [nextQueryInput, settingsCascade]
    )

    const onSelectCategory = useCallback(
        (category: ExtensionCategoryOrAll) => {
            const query = getQueryFromLocation(window.location)
            const currentCategory = getCategoryFromLocation(window.location)
            const experimental = getExperimentalFromLocation(window.location)

            if (category !== currentCategory) {
                setChangedCategory(true)
            }

            history.push(getRegistryLocationDescriptor({ query, category, experimental }))
        },
        [history]
    )

    // Keep state in sync with URL
    useEffect(() => {
        // kicks off initial request
        onQueryChangeImmediate()
    }, [location, onQueryChangeImmediate])

    const isLoading = !data || data === LOADING || changedCategory

    return (
        <>
            <div className="container">
                <PageTitle title="Extensions" />
                <div className="d-flex mt-3 pt-3">
                    <ExtensionRegistrySidenav
                        selectedCategory={selectedCategory}
                        onSelectCategory={onSelectCategory}
                        enablementFilter={enablementFilter}
                        setEnablementFilter={setEnablementFilter}
                        showExperimentalExtensions={showExperimentalExtensions}
                        toggleExperimentalExtensions={toggleExperimentalExtensions}
                    />
                    <div className="flex-grow-1">
                        <div className="mb-5">
                            <div className="row">
                                <span className="mb-3 col-lg-10">
                                    Connect all your other tools to get things like test coverage, 1-click open file in
                                    editor, custom highlighting, and information from your other favorite services all
                                    in one place on Sourcegraph.
                                </span>
                            </div>
                            <Form onSubmit={preventDefault}>
                                <Input
                                    data-testid="test-extension-registry-input"
                                    className="w-100 shadow"
                                    type="search"
                                    placeholder="Search extensions..."
                                    name="query"
                                    value={query}
                                    onChange={onQueryChangeEvent}
                                    autoFocus={true}
                                    autoComplete="off"
                                    autoCorrect="off"
                                    autoCapitalize="off"
                                    spellCheck={false}
                                />
                            </Form>
                            {!authenticatedUser && (
                                <Alert className="my-4" variant="info">
                                    <span>An account is required to create, enable and disable extensions. </span>
                                    <AlertLink to={buildGetStartedURL('extension-registry', '/extensions')}>
                                        Get started!
                                    </AlertLink>
                                </Alert>
                            )}
                            <ExtensionsList
                                {...props}
                                data={isLoading ? LOADING : data}
                                query={query}
                                enablementFilter={enablementFilter}
                                selectedCategory={selectedCategory}
                                onShowFullCategoryClicked={onSelectCategory}
                                showExperimentalExtensions={showExperimentalExtensions}
                            />
                        </div>
                        {/* Only show the banner when there are no selected categories and it is not loading */}
                        {selectedCategory === 'All' && !isLoading && (
                            <>
                                <hr className="mt-5" />
                                <div className="my-4 justify-content-center">
                                    <ExtensionBanner />
                                </div>
                            </>
                        )}
                        {props.isSourcegraphDotCom && (
                            <>
                                <hr className="mt-5" />
                                <div className="my-4 justify-content-center">
                                    You may use the Sourcegraph.com extension registry only with Sourcegraph{' '}
                                    <Link to="https://about.sourcegraph.com/pricing/">
                                        self-hosted or managed instances
                                    </Link>
                                    , <Link to="http://sourcegraph.com/">Sourcegraph.com</Link>, and Sourcegraph's{' '}
                                    <Link to="https://docs.sourcegraph.com/integration/browser_extension">
                                        browser extensions
                                    </Link>{' '}
                                    and{' '}
                                    <Link to="https://docs.sourcegraph.com/integration/editor">
                                        editor integrations
                                    </Link>
                                    . You may not use the Sourcegraph.com extension registry with Sourcegraph OSS. Learn
                                    more about the Sourcegraph.com extension registry and administration options in our{' '}
                                    <Link to="https://docs.sourcegraph.com/admin/extensions">
                                        extensions documentation
                                    </Link>
                                    .
                                </div>
                            </>
                        )}
                    </div>
                </div>
            </div>
        </>
    )
}

function getQueryFromLocation(location: Pick<H.Location, 'search'>): string {
    const parameters = new URLSearchParams(location.search)
    return parameters.get(URL_QUERY_PARAM) || ''
}

function getCategoryFromLocation(location: Pick<H.Location, 'search'>): ExtensionCategoryOrAll {
    const parameters = new URLSearchParams(location.search)
    const category = parameters.get(URL_CATEGORY_PARAM)

    if (category && isExtensionCategory(category)) {
        return category
    }

    return 'All'
}

function getExperimentalFromLocation(location: Pick<H.Location, 'search'>): boolean {
    const parameters = new URLSearchParams(location.search)
    const experimental = parameters.get('experimental')

    return experimental === 'true'
}

/**
 * Returns location descriptor object to push/replace onto the history stack
 * whenever the query or category is changed.
 */
function getRegistryLocationDescriptor({
    query,
    category,
    experimental,
}: {
    query: string
    category: ExtensionCategoryOrAll
    experimental: boolean
}): H.LocationDescriptorObject {
    const search = new URLSearchParams()

    search.set(URL_CATEGORY_PARAM, category)
    if (query) {
        search.set(URL_QUERY_PARAM, query)
    }
    if (experimental) {
        search.set('experimental', 'true')
    }

    return {
        search: search.toString(),
        hash: window.location.hash,
    }
}

function isExtensionCategory(category: string): category is ExtensionCategory {
    return EXTENSION_CATEGORIES.includes(category as ExtensionCategory)
}

function preventDefault(event: React.FormEvent): void {
    event.preventDefault()
}
