import * as H from 'history'
import React, { useEffect, useState, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { concat, of, timer } from 'rxjs'
import { debounce, delay, map, switchMap, takeUntil, tap, distinctUntilKeyChanged } from 'rxjs/operators'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { ConfiguredRegistryExtension } from '@sourcegraph/shared/src/extensions/extension'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { ExtensionCategory } from '@sourcegraph/shared/src/schema/extensionSchema'
import { Settings, SettingsCascadeProps, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { createAggregateError, ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useEventObservable } from '@sourcegraph/shared/src/util/useObservable'

import { PageTitle } from '../components/PageTitle'
import {
    RegistryExtensionsResult,
    RegistryExtensionFieldsForList,
    RegistryExtensionsVariables,
} from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'

import { ExtensionBanner } from './ExtensionBanner'
import { configureExtensionRegistry, ConfiguredExtensionRegistry } from './extensions'
import { ExtensionsAreaRouteContext } from './ExtensionsArea'
import { ExtensionsList } from './ExtensionsList'
import { ExtensionsQueryInputToolbar } from './ExtensionsQueryInputToolbar'

interface Props
    extends Pick<ExtensionsAreaRouteContext, 'authenticatedUser' | 'subject'>,
        PlatformContextProps<'settings' | 'updateSettings' | 'requestGraphQL'>,
        SettingsCascadeProps,
        ThemeProps {
    location: H.Location
    history: H.History
}

const LOADING = 'loading' as const
const URL_QUERY_PARAM = 'query'

export type ExtensionListData = typeof LOADING | (ConfiguredExtensionRegistry & { error: string | null }) | ErrorLike

export type ExtensionsEnablement = 'all' | 'enabled' | 'disabled'

const extensionRegistryQuery = gql`
    query RegistryExtensions($query: String, $prioritizeExtensionIDs: [String!]!) {
        extensionRegistry {
            extensions(query: $query, prioritizeExtensionIDs: $prioritizeExtensionIDs) {
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

export type ConfiguredExtensionCache = Map<
    string,
    Pick<ConfiguredRegistryExtension<RegistryExtensionFieldsForList>, 'manifest' | 'id'>
>

/** A page that displays overview information about the available extensions. */
export const ExtensionRegistry: React.FunctionComponent<Props> = props => {
    useEffect(() => eventLogger.logViewEvent('ExtensionsOverview'), [])

    const { history, location, settingsCascade, platformContext, authenticatedUser } = props

    // Update the cache after each response. This speeds up client-side filtering.
    // Lazy initialize cache ref. Don't mind the `useState` abuse:
    // - can't use `useMemo` here (https://github.com/facebook/react/issues/14490#issuecomment-454973512)
    // - `useRef` is just `useState`
    const [configuredExtensionCache] = useState<ConfiguredExtensionCache>(
        () => new Map<string, ConfiguredRegistryExtension<RegistryExtensionFieldsForList>>()
    )

    const [query, setQuery] = useState(getQueryFromProps(location))
    // Selected categories in order of selection. Used to assign extensions to their earliest selected category.
    const [selectedCategories, setSelectedCategories] = useState<ExtensionCategory[]>([])
    // Filter extensions by enablement state: enabled, disabled, or all.
    const [enablementFilter, setEnablementFilter] = useState<ExtensionsEnablement>('all')
    // Programming language extensions are hidden by default. Users cannot un-show PL extensions once toggled.
    const [showMoreExtensions, setShowMoreExtensions] = useState(false)

    /**
     * Note: pass `settingsCascade` instead of making it a dependency to prevent creating
     * new subscriptions when user toggles extensions
     */
    const [nextQueryInput, data] = useEventObservable<
        { query: string; immediate: boolean; settingsCascade: SettingsCascadeOrError<Settings> },
        ExtensionListData
    >(
        useCallback(
            newQueries =>
                newQueries.pipe(
                    distinctUntilKeyChanged('query'),
                    tap(({ query }) => {
                        setQuery(query)

                        history.replace({
                            search: query ? new URLSearchParams({ [URL_QUERY_PARAM]: query }).toString() : '',
                            hash: location.hash,
                        })
                    }),
                    debounce(({ immediate }) => timer(immediate ? 0 : 50)),
                    distinctUntilKeyChanged('query'),
                    switchMap(({ query, immediate, settingsCascade }) => {
                        let viewerConfiguredExtensions: string[] = []
                        if (!isErrorLike(settingsCascade.final)) {
                            if (settingsCascade.final?.extensions) {
                                viewerConfiguredExtensions = Object.keys(settingsCascade.final.extensions)
                            }
                        }
                        const resultOrError = platformContext.requestGraphQL<
                            RegistryExtensionsResult,
                            RegistryExtensionsVariables
                        >({
                            request: extensionRegistryQuery,
                            variables: { query, prioritizeExtensionIDs: viewerConfiguredExtensions },
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

                        return {
                            error,
                            ...configureExtensionRegistry(nodes, configuredExtensionCache),
                        }
                    })
                ),
            [platformContext, history, location.hash, configuredExtensionCache]
        )
    )

    const onQueryChangeEvent = useCallback(
        (event: React.FormEvent<HTMLInputElement>) =>
            nextQueryInput({ query: event.currentTarget.value, immediate: false, settingsCascade }),
        [nextQueryInput, settingsCascade]
    )

    const onQueryChangeImmediate = useCallback(
        (query: string) => nextQueryInput({ query, immediate: true, settingsCascade }),
        [nextQueryInput, settingsCascade]
    )

    useEffect(() => {
        // kicks off initial request
        onQueryChangeImmediate(getQueryFromProps(location))
    }, [location, onQueryChangeImmediate])

    const isLoading = !data || data === LOADING

    return (
        <>
            <div className="container">
                <PageTitle title="Extensions" />

                <div className="pt-3">
                    <div className="mb-5">
                        <Form onSubmit={preventDefault} className="form-inline">
                            <div className="shadow flex-grow-1 mb-2">
                                <input
                                    className="form-control w-100 test-extension-registry-input"
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
                            </div>
                        </Form>
                        <ExtensionsQueryInputToolbar
                            selectedCategories={selectedCategories}
                            onSelectCategories={setSelectedCategories}
                            enablementFilter={enablementFilter}
                            setEnablementFilter={setEnablementFilter}
                        />
                        {!authenticatedUser && (
                            <div className="alert alert-info my-4">
                                <span>An account is required to create, enable and disable extensions. </span>
                                <Link to="/sign-up?returnTo=/extensions">
                                    <span className="alert-link">Register now!</span>
                                </Link>
                            </div>
                        )}
                        <ExtensionsList
                            {...props}
                            data={data}
                            query={query}
                            enablementFilter={enablementFilter}
                            selectedCategories={selectedCategories}
                            showMoreExtensions={showMoreExtensions}
                        />
                    </div>
                    {!isLoading && !showMoreExtensions && selectedCategories.length === 0 && (
                        <div className="d-flex justify-content-center">
                            <button
                                type="button"
                                className="btn btn-outline-secondary"
                                onClick={() => setShowMoreExtensions(true)}
                            >
                                Show more extensions
                            </button>
                        </div>
                    )}
                </div>
            </div>
            {/* Only show the banner when there are no selected categories and it is not loading */}
            {selectedCategories.length === 0 && !isLoading && (
                <>
                    <hr className="mt-5" />
                    <div className="my-5 row justify-content-center">
                        <div className="mx-auto col-sm-12 col-md-8 col-lg-8 col-xl-6">
                            <ExtensionBanner />
                        </div>
                    </div>
                </>
            )}
        </>
    )
}

function getQueryFromProps(location: H.Location): string {
    const parameters = new URLSearchParams(location.search)
    return parameters.get(URL_QUERY_PARAM) || ''
}

function preventDefault(event: React.FormEvent): void {
    event.preventDefault()
}
