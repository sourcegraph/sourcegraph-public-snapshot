import classNames from 'classnames'
import React, { useCallback, useMemo, useState } from 'react'
import { Observable } from 'rxjs'
import { useDeepCompareEffectNoCheck } from 'use-deep-compare-effect'

import {
    SearchPatternType,
    fetchAutoDefinedSearchContexts,
    getUserSearchContextNamespaces,
    fetchSearchContexts,
} from '@sourcegraph/search'
import { SearchBox, StreamingProgress, StreamingSearchResultsList } from '@sourcegraph/search-ui'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { fetchHighlightedFileLineRanges } from '@sourcegraph/shared/src/backend/file'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { appendContextFilter, updateFilters } from '@sourcegraph/shared/src/search/query/transformer'
import { LATEST_VERSION, SearchMatch } from '@sourcegraph/shared/src/search/stream'
import { globbingEnabledFromSettings } from '@sourcegraph/shared/src/util/globbing'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { SearchResultsState } from '../../state'
import { WebviewPageProps } from '../platform/context'

import { SearchBetaIcon } from './components/icons'
import { SavedSearchCreateForm } from './components/SavedSearchForm'
import { SearchPageCta } from './components/SearchCta'
import { SearchResultsInfoBar } from './components/SearchResultsInfoBar'
import styles from './index.module.scss'

export interface SearchResultsViewProps extends WebviewPageProps {
    context: SearchResultsState['context']
}

export const SearchResultsView: React.FunctionComponent<SearchResultsViewProps> = ({
    extensionCoreAPI,
    authenticatedUser,
    platformContext,
    settingsCascade,
    theme,
    context,
    instanceURL,
}) => {
    const [userQueryState, setUserQueryState] = useState(context.submittedSearchQueryState.queryState)

    const [allExpanded, setAllExpanded] = useState(false)
    const onExpandAllResultsToggle = useCallback(() => {
        setAllExpanded(oldValue => !oldValue)
        platformContext.telemetryService.log(allExpanded ? 'allResultsExpanded' : 'allResultsCollapsed')
    }, [allExpanded, platformContext])

    const [showSavedSearchForm, setShowSavedSearchForm] = useState(false)

    // Update local query state on e.g. sidebar events.
    // TODO: create an API shared with SearchHomeView.

    useDeepCompareEffectNoCheck(() => {
        setUserQueryState(context.submittedSearchQueryState.queryState)
        // It's a whole new object on each state update, so we need
        // to compare (alternatively, construct full query TODO)
    }, [context.submittedSearchQueryState.queryState])

    const onSubmit = useCallback(
        (options?: { caseSensitive?: boolean; patternType?: SearchPatternType; newQuery?: string }) => {
            const previousSearchQueryState = context.submittedSearchQueryState

            extensionCoreAPI
                .streamSearch(options?.newQuery ?? userQueryState.query, {
                    caseSensitive: options?.caseSensitive ?? previousSearchQueryState.searchCaseSensitivity,
                    patternType: options?.patternType ?? previousSearchQueryState.searchPatternType,
                    version: LATEST_VERSION,
                    trace: undefined,
                })
                .then(() => {})
                .catch(() => {})
        },
        [userQueryState.query, context.submittedSearchQueryState, extensionCoreAPI]
    )

    // Submit new search on change
    const setCaseSensitivity = useCallback(
        (caseSensitive: boolean) => {
            onSubmit({ caseSensitive })
        },
        [onSubmit]
    )

    // Submit new search on change
    const setPatternType = useCallback(
        (patternType: SearchPatternType) => {
            console.log({ patternType })
            onSubmit({ patternType })
        },
        [onSubmit]
    )

    const fetchHighlightedFileLineRangesWithContext = useCallback(
        (parameters: FetchFileParameters) => fetchHighlightedFileLineRanges({ ...parameters, platformContext }),
        [platformContext]
    )

    const fetchStreamSuggestions = useCallback(
        (query): Observable<SearchMatch[]> =>
            wrapRemoteObservable(extensionCoreAPI.fetchStreamSuggestions(query, instanceURL)),
        [extensionCoreAPI, instanceURL]
    )

    const globbing = useMemo(() => globbingEnabledFromSettings(settingsCascade), [settingsCascade])

    const setSelectedSearchContextSpec = useCallback(
        (spec: string) => {
            extensionCoreAPI
                .setSelectedSearchContextSpec(spec)
                .catch(error => {
                    console.error('Error persisting search context spec.', error)
                })
                .finally(() => {
                    // Execute search with new context state
                    onSubmit()
                })
        },
        [extensionCoreAPI, onSubmit]
    )

    const onSearchAgain = useCallback(
        (additionalFilters: string[]) => {
            platformContext.telemetryService.log('SearchSkippedResultsAgainClicked')
            onSubmit({
                newQuery: applyAdditionalFilters(context.submittedSearchQueryState.queryState.query, additionalFilters),
            })
        },
        [context.submittedSearchQueryState.queryState, platformContext, onSubmit]
    )

    const onShareResultsClick = useCallback((): void => {
        const queryState = context.submittedSearchQueryState

        const path = `/search?${buildSearchURLQuery(
            queryState.queryState.query,
            queryState.searchPatternType,
            queryState.searchCaseSensitivity,
            context.selectedSearchContextSpec
        )}&utm_campaign=vscode-extension&utm_medium=direct_traffic&utm_source=vscode-extension&utm_content=save-search`
        extensionCoreAPI.copyLink(new URL(path, instanceURL).href).catch(error => {
            console.error('Error copying search link to clipboard:', error)
        })
    }, [context, instanceURL, extensionCoreAPI])

    const fullQuery = useMemo(
        () =>
            appendContextFilter(context.submittedSearchQueryState.queryState.query, context.selectedSearchContextSpec),
        [context]
    )

    const isSourcegraphDotCom = useMemo(() => {
        const hostname = new URL(instanceURL).hostname
        return hostname === 'sourcegraph.com' || hostname === 'www.sourcegraph.com'
    }, [instanceURL])

    const onSignUpClick = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            platformContext.telemetryService.log(
                'VSCESearchPageClicked',
                { campaign: 'Sign up link' },
                { campaign: 'Sign up link' }
            )
        },
        [platformContext.telemetryService]
    )

    return (
        <div className={styles.resultsViewLayout}>
            {/* eslint-disable-next-line react/forbid-elements */}
            <form
                className="d-flex w-100 my-2 pb-2 border-bottom"
                onSubmit={event => {
                    event.preventDefault()
                    onSubmit()
                }}
            >
                <SearchBox
                    caseSensitive={context.submittedSearchQueryState?.searchCaseSensitivity}
                    setCaseSensitivity={setCaseSensitivity}
                    patternType={context.submittedSearchQueryState?.searchPatternType}
                    setPatternType={setPatternType}
                    isSourcegraphDotCom={isSourcegraphDotCom}
                    hasUserAddedExternalServices={false}
                    hasUserAddedRepositories={true} // Used for search context CTA, which we won't show here.
                    structuralSearchDisabled={false}
                    queryState={userQueryState}
                    onChange={setUserQueryState}
                    onSubmit={onSubmit}
                    authenticatedUser={authenticatedUser}
                    searchContextsEnabled={true}
                    showSearchContext={true}
                    showSearchContextManagement={false}
                    defaultSearchContextSpec="global"
                    setSelectedSearchContextSpec={setSelectedSearchContextSpec}
                    selectedSearchContextSpec={context.selectedSearchContextSpec}
                    fetchSearchContexts={fetchSearchContexts}
                    fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                    getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                    fetchStreamSuggestions={fetchStreamSuggestions}
                    settingsCascade={settingsCascade}
                    globbing={globbing}
                    isLightTheme={theme === 'theme-light'}
                    telemetryService={platformContext.telemetryService}
                    platformContext={platformContext}
                    className={classNames('flex-grow-1 flex-shrink-past-contents', styles.searchBox)}
                    containerClassName={styles.searchBoxContainer}
                />
            </form>
            <div className={styles.resultsViewScrollContainer}>
                {!authenticatedUser && (
                    <SearchPageCta
                        icon={<SearchBetaIcon />}
                        ctaTitle="Sign up to add your public and private repositories and access other features"
                        ctaDescription="Do all the things editors can’t: search multiple repos & commit history, monitor, save searches and more."
                        buttonText="Create a free account"
                        onClickAction={onSignUpClick}
                    />
                )}
                <SearchResultsInfoBar
                    onShareResultsClick={onShareResultsClick}
                    showSavedSearchForm={showSavedSearchForm}
                    setShowSavedSearchForm={setShowSavedSearchForm}
                    extensionCoreAPI={extensionCoreAPI}
                    patternType={context.submittedSearchQueryState.searchPatternType}
                    authenticatedUser={authenticatedUser}
                    platformContext={platformContext}
                    stats={
                        <StreamingProgress
                            progress={context.searchResults?.progress || { durationMs: 0, matchCount: 0, skipped: [] }}
                            state={context.searchResults?.state || 'loading'}
                            onSearchAgain={onSearchAgain}
                            showTrace={false}
                        />
                    }
                    allExpanded={allExpanded}
                    onExpandAllResultsToggle={onExpandAllResultsToggle}
                />
                {authenticatedUser && showSavedSearchForm && (
                    <SavedSearchCreateForm
                        authenticatedUser={authenticatedUser}
                        submitLabel="Add saved search"
                        title="Add saved search"
                        fullQuery={fullQuery}
                        onComplete={() => setShowSavedSearchForm(false)}
                        platformContext={platformContext}
                        instanceURL={instanceURL}
                    />
                )}
                <StreamingSearchResultsList
                    isLightTheme={theme === 'theme-light'}
                    settingsCascade={settingsCascade}
                    telemetryService={platformContext.telemetryService}
                    allExpanded={allExpanded}
                    // Debt: dotcom prop used only for "run search" link
                    // for search examples. Fix on VSCE.
                    isSourcegraphDotCom={false}
                    searchContextsEnabled={true}
                    showSearchContext={true}
                    platformContext={platformContext}
                    results={context.searchResults ?? undefined}
                    authenticatedUser={authenticatedUser}
                    fetchHighlightedFileLineRanges={fetchHighlightedFileLineRangesWithContext}
                    executedQuery={context.submittedSearchQueryState.queryState.query}
                    resultClassName="mr-0"
                    // TODO "no results" video thumbnail assets
                    // In build, copy ui/assets/img folder to dist/
                    assetsRoot="https://raw.githubusercontent.com/sourcegraph/sourcegraph/main/ui/assets"
                />
            </div>
        </div>
    )
}

const applyAdditionalFilters = (query: string, additionalFilters: string[]): string => {
    let newQuery = query
    for (const filter of additionalFilters) {
        const fieldValue = filter.split(':', 2)
        newQuery = updateFilters(newQuery, fieldValue[0], fieldValue[1])
    }
    return newQuery
}
