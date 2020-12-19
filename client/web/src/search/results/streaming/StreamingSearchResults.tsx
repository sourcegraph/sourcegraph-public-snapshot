import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import FileIcon from 'mdi-react/FileIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Observable } from 'rxjs'
import { FetchFileParameters } from '../../../../../shared/src/components/CodeExcerpt'
import { FileMatch } from '../../../../../shared/src/components/FileMatch'
import { VirtualList } from '../../../../../shared/src/components/VirtualList'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { SearchPatternType } from '../../../../../shared/src/graphql-operations'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { VersionContextProps } from '../../../../../shared/src/search/util'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { ThemeProps } from '../../../../../shared/src/theme'
import { isDefined } from '../../../../../shared/src/util/types'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { AuthenticatedUser } from '../../../auth'
import { PageTitle } from '../../../components/PageTitle'
import { SearchResult } from '../../../components/SearchResult'
import { SavedSearchModal } from '../../../savedSearches/SavedSearchModal'
import { VersionContext } from '../../../schema/site.schema'
import { QueryState, submitSearch } from '../../helpers'
import { SearchAlert } from '../SearchAlert'
import { LATEST_VERSION } from '../SearchResults'
import { SearchResultsInfoBar } from '../SearchResultsInfoBar'
import { SearchResultTypeTabs } from '../SearchResultTypeTabs'
import { VersionContextWarning } from '../VersionContextWarning'
import { StreamingProgress } from './progress/StreamingProgress'
import { StreamingSearchResultsFilterBars } from './StreamingSearchResultsFilterBars'
import {
    CaseSensitivityProps,
    parseSearchURL,
    PatternTypeProps,
    SearchStreamingProps,
    resolveVersionContext,
} from '../..'
import { ErrorAlert } from '../../../components/alerts'
import { eventLogger } from '../../../tracking/eventLogger'

export interface StreamingSearchResultsProps
    extends SearchStreamingProps,
        PatternTypeProps,
        VersionContextProps,
        CaseSensitivityProps,
        SettingsCascadeProps,
        ExtensionsControllerProps<'executeCommand' | 'extHostAPI' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        TelemetryProps,
        ThemeProps {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    navbarSearchQueryState: QueryState

    setVersionContext: (versionContext: string | undefined) => void
    availableVersionContexts: VersionContext[] | undefined
    previousVersionContext: string | null

    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

const initialItemsToShow = 15
const incrementalItemsToShow = 10

export const StreamingSearchResults: React.FunctionComponent<StreamingSearchResultsProps> = props => {
    const {
        patternType: currentPatternType,
        setPatternType,
        caseSensitive: currentCaseSensitive,
        setCaseSensitivity,
        versionContext: currentVersionContext,
        setVersionContext,
        streamSearch,
        location,
        history,
        availableVersionContexts,
        previousVersionContext,
        authenticatedUser,
    } = props

    const { query = '', patternType, caseSensitive, versionContext } = parseSearchURL(props.location.search)

    useEffect(() => {
        if (patternType && patternType !== currentPatternType) {
            setPatternType(patternType)
        }
    }, [patternType, currentPatternType, setPatternType])

    useEffect(() => {
        if (caseSensitive && caseSensitive !== currentCaseSensitive) {
            setCaseSensitivity(caseSensitive)
        }
    }, [caseSensitive, currentCaseSensitive, setCaseSensitivity])

    useEffect(() => {
        const resolvedContext = resolveVersionContext(versionContext, availableVersionContexts)
        if (resolvedContext !== currentVersionContext) {
            setVersionContext(resolvedContext)
        }
    }, [versionContext, currentVersionContext, setVersionContext, availableVersionContexts])

    const results = useObservable(
        useMemo(
            () =>
                streamSearch(
                    query,
                    LATEST_VERSION,
                    patternType ?? SearchPatternType.literal,
                    resolveVersionContext(versionContext, availableVersionContexts)
                ),
            [streamSearch, query, patternType, versionContext, availableVersionContexts]
        )
    )

    const [allExpanded, setAllExpanded] = useState(false)
    const onExpandAllResultsToggle = useCallback(() => setAllExpanded(oldValue => !oldValue), [setAllExpanded])

    const [showSavedSearchModal, setShowSavedSearchModal] = useState(false)
    const onSaveQueryClick = useCallback(() => setShowSavedSearchModal(true), [])
    const onSaveQueryModalClose = useCallback(() => setShowSavedSearchModal(false), [])

    const [showVersionContextWarning, setShowVersionContextWarning] = useState(false)
    useEffect(
        () => {
            const searchParameters = new URLSearchParams(location.search)
            const versionFromURL = searchParameters.get('c')

            if (searchParameters.has('from-context-toggle')) {
                // The query param `from-context-toggle` indicates that the version context
                // changed from the version context toggle. In this case, we don't warn
                // users that the version context has changed.
                searchParameters.delete('from-context-toggle')
                history.replace({
                    search: searchParameters.toString(),
                    hash: history.location.hash,
                })
                setShowVersionContextWarning(false)
            } else {
                setShowVersionContextWarning(
                    (availableVersionContexts && versionFromURL !== previousVersionContext) || false
                )
            }
        },
        // Only show warning when URL changes
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [location.search]
    )
    const onDismissVersionContextWarning = useCallback(() => setShowVersionContextWarning(false), [
        setShowVersionContextWarning,
    ])

    const [itemsToShow, setItemsToShow] = useState(initialItemsToShow)
    const onBottomHit = useCallback(
        () => setItemsToShow(items => Math.min(results?.results.length || 0, items + incrementalItemsToShow)),
        [results?.results.length]
    )
    const logSearchResultClicked = useCallback(() => props.telemetryService.log('SearchResultClicked'), [
        props.telemetryService,
    ])
    const renderResult = (result: GQL.GenericSearchResultInterface | GQL.IFileMatch): JSX.Element | undefined => {
        switch (result.__typename) {
            case 'FileMatch':
                return (
                    <FileMatch
                        key={'file:' + result.file.url}
                        location={location}
                        eventLogger={eventLogger}
                        icon={result.lineMatches && result.lineMatches.length > 0 ? SourceRepositoryIcon : FileIcon}
                        result={result}
                        onSelect={logSearchResultClicked}
                        expanded={false}
                        showAllMatches={false}
                        isLightTheme={props.isLightTheme}
                        allExpanded={allExpanded}
                        fetchHighlightedFileLineRanges={props.fetchHighlightedFileLineRanges}
                        settingsCascade={props.settingsCascade}
                    />
                )
        }
        return (
            <SearchResult key={result.url} result={result} isLightTheme={props.isLightTheme} history={props.history} />
        )
    }

    const onSearchAgain = useCallback(
        (additionalFilters: string[]) => {
            const newQuery = [query, ...additionalFilters].join(' ')
            submitSearch({ ...props, query: newQuery, source: 'excludedResults' })
        },
        [query, props]
    )

    return (
        <div className="test-search-results search-results d-flex flex-column w-100">
            <PageTitle key="page-title" title={query} />
            <StreamingSearchResultsFilterBars {...props} results={results} />
            <div className="search-results-list">
                <div className="d-lg-flex mb-2 align-items-end flex-wrap">
                    <SearchResultTypeTabs
                        {...props}
                        query={props.navbarSearchQueryState.query}
                        className="search-results-list__tabs"
                    />

                    <SearchResultsInfoBar
                        {...props}
                        query={query}
                        resultsFound={results ? results.results.length > 0 : false}
                        className="border-bottom flex-grow-1"
                        allExpanded={allExpanded}
                        onExpandAllResultsToggle={onExpandAllResultsToggle}
                        onSaveQueryClick={onSaveQueryClick}
                        stats={
                            <StreamingProgress
                                progress={results?.progress || { durationMs: 0, matchCount: 0, skipped: [] }}
                                state={results?.state || 'loading'}
                                onSearchAgain={onSearchAgain}
                            />
                        }
                    />
                </div>

                {showVersionContextWarning && (
                    <VersionContextWarning
                        versionContext={currentVersionContext}
                        onDismissWarning={onDismissVersionContextWarning}
                    />
                )}

                {showSavedSearchModal && (
                    <SavedSearchModal
                        {...props}
                        query={query}
                        authenticatedUser={authenticatedUser}
                        onDidCancel={onSaveQueryModalClose}
                    />
                )}

                {results?.alert && (
                    <SearchAlert
                        alert={results.alert}
                        caseSensitive={caseSensitive}
                        patternType={patternType}
                        versionContext={versionContext}
                    />
                )}

                {/* Results */}
                <VirtualList
                    className="mt-2"
                    itemsToShow={itemsToShow}
                    onShowMoreItems={onBottomHit}
                    items={results?.results.map(result => renderResult(result)).filter(isDefined) || []}
                />

                {(!results || results?.state === 'loading') && (
                    <div className="text-center my-4" data-testid="loading-container">
                        <LoadingSpinner className="icon-inline" />
                    </div>
                )}

                {results?.state === 'error' && (
                    <ErrorAlert
                        className="m-2"
                        data-testid="search-results-list-error"
                        error={results.error}
                        history={history}
                    />
                )}

                {results?.state === 'complete' && !results?.alert && results?.results.length === 0 && (
                    <div className="alert alert-info d-flex m-2">
                        <h3 className="m-0">
                            <SearchIcon className="icon-inline" /> No results
                        </h3>
                    </div>
                )}

                {results?.state === 'complete' && results?.results.length > 0 && (
                    <small className="d-block my-4 text-center">Showing {results?.results.length} results</small>
                )}
            </div>
        </div>
    )
}
