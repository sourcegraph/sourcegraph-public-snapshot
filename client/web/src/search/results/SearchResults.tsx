import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { concat, from, Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, startWith, switchMap, tap } from 'rxjs/operators'
import {
    parseSearchURLQuery,
    parseSearchURLPatternType,
    PatternTypeProps,
    CaseSensitivityProps,
    parseSearchURL,
    resolveVersionContext,
    MutableVersionContextProps,
    ParsedSearchQueryProps,
    SearchContextProps,
} from '..'
import { Contributions, Evaluated } from '../../../../shared/src/api/protocol'
import { FetchFileParameters } from '../../../../shared/src/components/CodeExcerpt'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { isSettingsValid, SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { ErrorLike, isErrorLike, asError } from '../../../../shared/src/util/errors'
import { PageTitle } from '../../components/PageTitle'
import { Settings } from '../../schema/settings.schema'
import { ThemeProps } from '../../../../shared/src/theme'
import { eventLogger, EventLogger } from '../../tracking/eventLogger'
import { isSearchResults, submitSearch, toggleSearchFilter, getSearchTypeFromQuery, QueryState } from '../helpers'
import { queryTelemetryData } from '../queryTelemetry'
import { DynamicSearchFilter, SearchResultsFilterBars } from './SearchResultsFilterBars'
import { SearchResultsList } from './SearchResultsList'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { Remote } from 'comlink'
import { FlatExtensionHostAPI } from '../../../../shared/src/api/contract'
import { DeployType } from '../../jscontext'
import { AuthenticatedUser } from '../../auth'
import { SearchPatternType } from '../../../../shared/src/graphql-operations'
import { shouldDisplayPerformanceWarning } from '../backend'
import { VersionContextWarning } from './VersionContextWarning'
import { CodeMonitoringProps } from '../../enterprise/code-monitoring'
import { wrapRemoteObservable } from '../../../../shared/src/api/client/api/common'

export interface SearchResultsProps
    extends ExtensionsControllerProps<'executeCommand' | 'extHostAPI' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        SettingsCascadeProps,
        TelemetryProps,
        ThemeProps,
        Pick<ParsedSearchQueryProps, 'parsedSearchQuery'>,
        PatternTypeProps,
        CaseSensitivityProps,
        MutableVersionContextProps,
        Pick<CodeMonitoringProps, 'enableCodeMonitoring'>,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    navbarSearchQueryState: QueryState
    telemetryService: Pick<EventLogger, 'log' | 'logViewEvent'>
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    searchRequest: (
        query: string,
        version: string,
        patternType: SearchPatternType,
        versionContext: string | undefined,
        extensionHostPromise: Promise<Remote<FlatExtensionHostAPI>>
    ) => Observable<GQL.ISearchResults | ErrorLike>
    isSourcegraphDotCom: boolean
    deployType: DeployType
}

interface SearchResultsState {
    /** The loaded search results, error or undefined while loading */
    resultsOrError?: GQL.ISearchResults
    allExpanded: boolean

    /* The time when loading the search results started. */
    loadingStarted: number

    // Saved Queries
    showSavedQueryModal: boolean

    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Evaluated<Contributions>

    /** Whether to show a warning saying that the URL has changed the version context. */
    showVersionContextWarning: boolean

    /** Whether the user has dismissed the version context warning. */
    dismissedVersionContextWarning?: boolean
}

/** All values that are valid for the `type:` filter. `null` represents default code search. */
export type SearchType = 'diff' | 'commit' | 'symbol' | 'repo' | 'path' | null

// The latest supported version of our search syntax. Users should never be able to determine the search version.
// The version is set based on the release tag of the instance. Anything before 3.9.0 will not pass a version parameter,
// and will therefore default to V1.
export const LATEST_VERSION = 'V2'

export class SearchResults extends React.Component<SearchResultsProps, SearchResultsState> {
    public state: SearchResultsState = {
        showSavedQueryModal: false,
        allExpanded: false,
        showVersionContextWarning: false,
        loadingStarted: 0,
    }
    /** Emits on componentDidUpdate with the new props */
    private componentUpdates = new Subject<SearchResultsProps>()

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const patternType = parseSearchURLPatternType(this.props.location.search)

        if (!patternType) {
            // If the patternType query parameter does not exist in the URL or is invalid, redirect to a URL which
            // has patternType=regexp appended. This is to ensure old URLs before requiring patternType still work.
            const query = parseSearchURLQuery(this.props.location.search) ?? ''
            const newLocation =
                '/search?' +
                buildSearchURLQuery(
                    query,
                    SearchPatternType.regexp,
                    this.props.caseSensitive,
                    this.props.versionContext,
                    this.props.selectedSearchContextSpec
                )
            this.props.history.replace(newLocation)
        }

        this.props.telemetryService.logViewEvent('SearchResults')

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(this.props),
                    map(props =>
                        parseSearchURL(props.location.search, {
                            appendCaseFilter: true,
                        })
                    ),
                    // Search when a new search query was specified in the URL
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    filter(
                        (
                            queryAndPatternTypeAndCase
                        ): queryAndPatternTypeAndCase is {
                            query: string
                            patternType: SearchPatternType
                            caseSensitive: boolean
                            versionContext: string | undefined
                        } => !!queryAndPatternTypeAndCase.query && !!queryAndPatternTypeAndCase.patternType
                    ),
                    tap(({ query, caseSensitive }) => {
                        const query_data = queryTelemetryData(query, caseSensitive)
                        this.props.telemetryService.log('SearchResultsQueried', {
                            code_search: { query_data },
                        })
                        if (query_data.query?.field_type && query_data.query.field_type.value_diff > 0) {
                            this.props.telemetryService.log('DiffSearchResultsQueried')
                        }
                    }),
                    switchMap(({ query, patternType, caseSensitive, versionContext }) =>
                        concat(
                            // Reset view state
                            [
                                {
                                    resultsOrError: undefined,
                                    loadingStarted: Date.now(),
                                    didSave: false,
                                    activeType: getSearchTypeFromQuery(query),
                                },
                            ],
                            // Do async search request
                            this.props
                                .searchRequest(
                                    query,
                                    LATEST_VERSION,
                                    patternType,
                                    resolveVersionContext(versionContext, this.props.availableVersionContexts),
                                    this.props.extensionsController.extHostAPI
                                )
                                .pipe(
                                    // Log telemetry
                                    tap(
                                        results => {
                                            this.props.telemetryService.log('SearchResultsFetched', {
                                                code_search: {
                                                    // 🚨 PRIVACY: never provide any private data in { code_search: { results } }.
                                                    results: {
                                                        results_count: isErrorLike(results)
                                                            ? 0
                                                            : results.results.length,
                                                        any_cloning: isErrorLike(results)
                                                            ? false
                                                            : results.cloning.length > 0,
                                                    },
                                                },
                                            })
                                            if (patternType && patternType !== this.props.patternType) {
                                                this.props.setPatternType(patternType)
                                            }
                                            if (caseSensitive !== this.props.caseSensitive) {
                                                this.props.setCaseSensitivity(caseSensitive)
                                            }

                                            this.props.setVersionContext(versionContext).catch(error => {
                                                console.error(
                                                    'Error sending initial versionContext to extensions',
                                                    error
                                                )
                                            })
                                        },
                                        error => {
                                            this.props.telemetryService.log('SearchResultsFetchFailed', {
                                                code_search: { error_message: asError(error).message },
                                            })
                                            console.error(error)
                                        }
                                    ),
                                    // Update view with results or error
                                    map(resultsOrError => ({ resultsOrError })),
                                    catchError(error => [{ resultsOrError: error }])
                                )
                        )
                    )
                )
                .subscribe(
                    newState => this.setState(newState as SearchResultsState),
                    error => console.error(error)
                )
        )

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(this.props),
                    distinctUntilChanged((a, b) => isEqual(a.location, b.location))
                )
                .subscribe(props => {
                    const searchParameters = new URLSearchParams(props.location.search)
                    const versionFromURL = searchParameters.get('c')

                    if (searchParameters.has('from-context-toggle')) {
                        // The query param `from-context-toggle` indicates that the version context
                        // changed from the version context toggle. In this case, we don't warn
                        // users that the version context has changed.
                        searchParameters.delete('from-context-toggle')
                        this.props.history.replace({
                            search: searchParameters.toString(),
                            hash: this.props.history.location.hash,
                        })
                        this.setState({ showVersionContextWarning: false })
                    } else {
                        this.setState({
                            showVersionContextWarning:
                                (props.availableVersionContexts && versionFromURL !== props.previousVersionContext) ||
                                false,
                        })
                    }
                })
        )

        this.subscriptions.add(
            from(this.props.extensionsController.extHostAPI)
                .pipe(switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getContributions())))
                .subscribe(contributions => this.setState({ contributions }))
        )
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private showSaveQueryModal = (): void => {
        this.setState({ showSavedQueryModal: true })
    }

    private onModalClose = (): void => {
        this.props.telemetryService.log('SavedQueriesToggleCreating', { queries: { creating: false } })
        this.setState({ showSavedQueryModal: false })
    }

    private onDismissWarning = (): void => {
        this.setState({ showVersionContextWarning: false })
    }

    private onFirstResultLoad = (): void => {
        const patternType = parseSearchURLPatternType(this.props.location.search)
        eventLogger.log(`search.latencies.frontend.${patternType || 'unknown'}.first-result`, {
            durationMs: Date.now() - this.state.loadingStarted,
        })
    }

    public render(): JSX.Element | null {
        const query = parseSearchURLQuery(this.props.location.search)
        const filters = this.getFilters()
        const repoFilters = this.getRepoFilters()
        const extensionFilters = this.state.contributions?.searchFilters

        const quickLinks =
            (isSettingsValid<Settings>(this.props.settingsCascade) && this.props.settingsCascade.final.quicklinks) || []

        return (
            <div className="test-search-results search-results d-flex flex-column w-100">
                <PageTitle key="page-title" title={query} />
                <SearchResultsFilterBars
                    navbarSearchQuery={this.props.navbarSearchQueryState.query}
                    searchSucceeded={isSearchResults(this.state.resultsOrError)}
                    resultsLimitHit={isSearchResults(this.state.resultsOrError) && this.state.resultsOrError.limitHit}
                    genericFilters={filters}
                    extensionFilters={extensionFilters}
                    repoFilters={repoFilters}
                    quickLinks={quickLinks}
                    onFilterClick={this.onDynamicFilterClicked}
                    onShowMoreResultsClick={this.showMoreResults}
                    calculateShowMoreResultsCount={this.calculateCount}
                />
                {this.state.showVersionContextWarning && (
                    <VersionContextWarning
                        versionContext={this.props.versionContext}
                        onDismissWarning={this.onDismissWarning}
                    />
                )}
                <SearchResultsList
                    {...this.props}
                    resultsOrError={this.state.resultsOrError}
                    onFirstResultLoad={this.onFirstResultLoad}
                    onShowMoreResultsClick={this.showMoreResults}
                    onExpandAllResultsToggle={this.onExpandAllResultsToggle}
                    allExpanded={this.state.allExpanded}
                    showSavedQueryModal={this.state.showSavedQueryModal}
                    onSaveQueryClick={this.showSaveQueryModal}
                    onSavedQueryModalClose={this.onModalClose}
                    shouldDisplayPerformanceWarning={shouldDisplayPerformanceWarning}
                />
            </div>
        )
    }

    /** Combines dynamic filters and search scopes into a list de-duplicated by value. */
    private getFilters(): DynamicSearchFilter[] {
        const filters = new Map<string, DynamicSearchFilter>()

        if (isSearchResults(this.state.resultsOrError) && this.state.resultsOrError.dynamicFilters) {
            let dynamicFilters = this.state.resultsOrError.dynamicFilters
            dynamicFilters = this.state.resultsOrError.dynamicFilters.filter(filter => filter.kind !== 'repo')
            for (const filter of dynamicFilters) {
                filters.set(filter.value, filter)
            }
        }
        const scopes =
            (isSettingsValid<Settings>(this.props.settingsCascade) &&
                this.props.settingsCascade.final['search.scopes']) ||
            []
        if (isSearchResults(this.state.resultsOrError) && this.state.resultsOrError.dynamicFilters) {
            for (const scope of scopes) {
                if (!filters.has(scope.value)) {
                    filters.set(scope.value, scope)
                }
            }
        } else {
            for (const scope of scopes) {
                // Check for if filter.value already exists and if so, overwrite with user's configured scope name.
                const existingFilter = filters.get(scope.value)
                // This works because user setting configs are the last to be processed after Global and Org.
                // Thus, user set filters overwrite the equal valued existing filters.
                if (existingFilter) {
                    existingFilter.name = scope.name || scope.value
                }
                filters.set(scope.value, existingFilter || scope)
            }
        }

        return [...filters.values()]
    }

    private getRepoFilters(): DynamicSearchFilter[] | undefined {
        if (isSearchResults(this.state.resultsOrError) && this.state.resultsOrError.dynamicFilters) {
            return this.state.resultsOrError.dynamicFilters
                .filter(filter => filter.kind === 'repo' && filter.value !== '')
                .map(filter => ({
                    name: filter.label,
                    value: filter.value,
                    count: filter.count,
                    limitHit: filter.limitHit,
                }))
        }
        return undefined
    }

    private showMoreResults = (): void => {
        // Requery with an increased max result count.
        const parameters = new URLSearchParams(this.props.location.search)
        let query = parameters.get('q') || ''

        const count = this.calculateCount()
        if (/count:(\d+)/.test(query)) {
            query = query.replace(/count:\d+/g, '').trim() + ` count:${count}`
        } else {
            query = `${query} count:${count}`
        }
        parameters.set('q', query)
        this.props.history.replace({ search: parameters.toString() })
    }

    private calculateCount = (): number => {
        // This function can only get called if the results were successfully loaded,
        // so casting is the right thing to do here
        const results = this.state.resultsOrError as GQL.ISearchResults

        const parameters = new URLSearchParams(this.props.location.search)
        const query = parameters.get('q') || ''

        if (/count:(\d+)/.test(query)) {
            return Math.max(results.matchCount * 2, 1000)
        }
        return Math.max(results.matchCount * 2 || 0, 1000)
    }

    private onExpandAllResultsToggle = (): void => {
        this.setState(
            state => ({ allExpanded: !state.allExpanded }),
            () => {
                this.props.telemetryService.log(this.state.allExpanded ? 'allResultsExpanded' : 'allResultsCollapsed')
            }
        )
    }

    private onDynamicFilterClicked = (value: string): void => {
        this.props.telemetryService.log('DynamicFilterClicked', {
            search_filter: { value },
        })

        const newQuery = toggleSearchFilter(this.props.navbarSearchQueryState.query, value)

        submitSearch({ ...this.props, query: newQuery, source: 'filter' })
    }
}
