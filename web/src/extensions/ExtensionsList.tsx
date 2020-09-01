import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import React, { useEffect, useState, useCallback, useRef } from 'react'
import { concat, of, timer } from 'rxjs'
import { debounce, delay, map, switchMap, takeUntil, tap, distinctUntilKeyChanged } from 'rxjs/operators'
import { ConfiguredRegistryExtension, isExtensionEnabled } from '../../../shared/src/extensions/extension'
import { gql } from '../../../shared/src/graphql/graphql'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import {
    Settings,
    SettingsCascadeProps,
    SettingsSubject,
    SettingsCascadeOrError,
} from '../../../shared/src/settings/settings'
import { createAggregateError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { Form } from '../components/Form'
import { extensionsQuery, isExtensionAdded } from './extension/extension'
import { ExtensionCard } from './ExtensionCard'
import { ExtensionsQueryInputToolbar } from './ExtensionsQueryInputToolbar'
import { ErrorAlert } from '../components/alerts'
import { useEventObservable } from '../../../shared/src/util/useObservable'
import {
    RegistryExtensionsResult,
    RegistryExtensionFieldsForList,
    RegistryExtensionsVariables,
} from '../graphql-operations'
import { categorizeExtensionRegistry, CategorizedExtensionRegistry, applyExtensionsEnablement } from './extensions'
import { ExtensionCategory, EXTENSION_CATEGORIES } from '../../../shared/src/schema/extensionSchema'
import { ShowMoreExtensions } from './ShowMoreExtensions'
import { ExtensionBanner } from './ExtensionBanner'
import { Link } from 'react-router-dom'
import { ExtensionsAreaRouteContext } from './ExtensionsArea'

interface Props
    extends SettingsCascadeProps,
        PlatformContextProps<'settings' | 'updateSettings' | 'requestGraphQL'>,
        Pick<ExtensionsAreaRouteContext, 'authenticatedUser'> {
    subject: Pick<SettingsSubject, 'id' | 'viewerCanAdminister'>
    location: H.Location
    history: H.History
}

const LOADING = 'loading' as const
const URL_QUERY_PARAM = 'query'

type NewExtensionListData = typeof LOADING | (CategorizedExtensionRegistry & { error: string | null }) | ErrorLike

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

/**
 * Displays a list of extensions.
 */
export const ExtensionsList: React.FunctionComponent<Props> = ({
    history,
    location,
    subject,
    settingsCascade,
    platformContext,
    authenticatedUser,
}) => {
    const { current: configuredExtensionCache } = useRef(
        new Map<string, ConfiguredRegistryExtension<RegistryExtensionFieldsForList>>()
    )
    const [query, setQuery] = useState(getQueryFromProps(location))
    const [selectedCategories, setSelectedCategories] = useState<ExtensionCategory[]>([])
    const [enablementFilter, setEnablementFilter] = useState<ExtensionsEnablement>('all')
    const [showMoreExtensions, setShowMoreExtensions] = useState(false)

    /**
     * Note: pass `settingsCascade` instead of making it a dependency to prevent creating
     * new subscriptions when user toggles extensions
     */
    const [nextQueryInput, data] = useEventObservable<
        { query: string; immediate: boolean; settingsCascade: SettingsCascadeOrError<Settings> },
        NewExtensionListData
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

                        if (!data || !data.extensionRegistry || !data.extensionRegistry.extensions) {
                            return createAggregateError(errors)
                        }

                        const { error, nodes } = data.extensionRegistry.extensions

                        return {
                            error,
                            ...categorizeExtensionRegistry(nodes, configuredExtensionCache),
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

    /**
     * Helper function to enable rendering logic outside of JSX without boilerplate
     *
     * TODO: extract to layout component
     */
    function registryLayout(content: JSX.Element, showShowMore: boolean = false): React.ReactElement {
        return (
            <>
                <div className="extensions-list">
                    <p>Improve your workflow with code intelligence, test coverage, and other useful information.</p>
                    <Form onSubmit={preventDefault} className="form-inline">
                        <input
                            className="form-control flex-grow-1 mr-1 mb-2 test-extension-registry-input extensions-list__input"
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

                    <ExtensionsQueryInputToolbar
                        query={query}
                        onQueryChange={onQueryChangeImmediate}
                        selectedCategories={selectedCategories}
                        setSelectedCategories={setSelectedCategories}
                        enablementFilter={enablementFilter}
                        setEnablementFilter={setEnablementFilter}
                    />

                    {!authenticatedUser && (
                        <div className="alert alert-info my-4">
                            <span>An account is required to create and configure extensions. </span>
                            <Link to="/sign-in?returnTo=/extensions">
                                <span className="alert-link">Register Now!</span>
                            </Link>
                        </div>
                    )}

                    {content}
                </div>

                {showShowMore && <ShowMoreExtensions setShowMoreExtensions={setShowMoreExtensions} />}

                {/* don't show banner when loading */}
                {data && data !== LOADING && <ExtensionBanner />}
            </>
        )
    }

    // no data on initial render
    if (!data || data === LOADING) {
        return registryLayout(<LoadingSpinner className="icon-inline" />)
    }

    if (isErrorLike(data)) {
        return registryLayout(<ErrorAlert error={data} history={history} />)
    }

    const { error, categories, extensions } = data

    if (Object.keys(extensions).length === 0) {
        return registryLayout(
            <>
                {error && <ErrorAlert className="mb-2" error={error} history={history} />}
                {query ? (
                    <div className="text-muted">
                        No extensions match <strong>{query}</strong>.
                    </div>
                ) : (
                    <span className="text-muted">No extensions found</span>
                )}
            </>
        )
    }

    /**
     * Apply categories:
     * - if user opts into seeing language extensions ('See more extensions') and no categories are selected, render all categories
     * - else if categories are selected, render the selected categories
     * - else, only render languages if selected
     */
    const renderLanguages =
        (selectedCategories.length === 0 && showMoreExtensions) || selectedCategories.includes('Programming languages')

    const filteredCategoryIDs = EXTENSION_CATEGORIES.filter(category => {
        if (category === 'Programming languages') {
            return renderLanguages
        }

        return selectedCategories.length === 0 || selectedCategories.includes(category)
    })

    // Apply enablement filter
    const filteredCategories = applyExtensionsEnablement(
        categories,
        filteredCategoryIDs,
        enablementFilter,
        settingsCascade.final
    )

    const categorySections: JSX.Element[] = []

    for (const category of filteredCategoryIDs) {
        // Don't render the section (including header) if no extensions belong to the category
        if (filteredCategories[category].length > 0) {
            categorySections.push(
                <div key={category} className="mt-1">
                    <h3 className="extensions-list__category">{category}</h3>
                    <div className="extensions-list__cards mt-1">
                        {filteredCategories[category].map(extensionId => (
                            <ExtensionCard
                                key={extensionId}
                                subject={subject}
                                extension={extensions[extensionId]}
                                settingsCascade={settingsCascade}
                                platformContext={platformContext}
                                enabled={isExtensionEnabled(settingsCascade.final, extensionId)}
                            />
                        ))}
                    </div>
                </div>
            )
        }
    }

    return registryLayout(
        <>
            {error && <ErrorAlert className="mb-2" error={error} history={history} />}
            {categorySections.length > 0 ? (
                categorySections
            ) : (
                <div className="text-muted">
                    No extensions match <strong>{query}</strong> in the selected categories.
                </div>
            )}
        </>,
        !showMoreExtensions && selectedCategories.length === 0
    )
}

function getQueryFromProps(location: H.Location): string {
    const parameters = new URLSearchParams(location.search)
    return parameters.get(URL_QUERY_PARAM) || ''
}

function preventDefault(event: React.FormEvent): void {
    event.preventDefault()
}

/**
 * Applies the query's client-side extensions search keywords #installed, #enabled, and #disabled by filtering
 * {@link registryExtensions}.
 *
 * @internal Exported for testing only.
 */
export function applyExtensionsQuery<X extends { extensionID: string }>(
    query: string,
    settings: Pick<Settings, 'extensions'>,
    registryExtensions: X[]
): X[] {
    const installed = query.includes(extensionsQuery({ installed: true }))
    const enabled = query.includes(extensionsQuery({ enabled: true }))
    const disabled = query.includes(extensionsQuery({ disabled: true }))
    return registryExtensions.filter(
        extension =>
            (!installed || isExtensionAdded(settings, extension.extensionID)) &&
            (!enabled || isExtensionEnabled(settings, extension.extensionID)) &&
            (!disabled ||
                (isExtensionAdded(settings, extension.extensionID) &&
                    !isExtensionEnabled(settings, extension.extensionID)))
    )
}
