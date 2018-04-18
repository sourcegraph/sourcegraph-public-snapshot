import DocumentIcon from '@sourcegraph/icons/lib/Document'
import Loader from '@sourcegraph/icons/lib/Loader'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as H from 'history'
import { upperFirst } from 'lodash-es'
import * as React from 'react'
import { FileMatch } from '../../components/FileMatch'
import { ModalContainer } from '../../components/ModalContainer'
import { eventLogger } from '../../tracking/eventLogger'
import { ErrorLike, isErrorLike } from '../../util/errors'
import { isSearchResults } from '../helpers'
import { parseSearchURLQuery } from '../index'
import { SavedQueryCreateForm } from '../saved-queries/SavedQueryCreateForm'
import { CommitSearchResult } from './CommitSearchResult'
import { RepositorySearchResult } from './RepositorySearchResult'
import { SearchAlert } from './SearchAlert'
import { SearchResultsInfoBar } from './SearchResultsInfoBar'

const DATA_CENTER_UPGRADE_STRING =
    'Upgrade to Sourcegraph Data Center for distributed on-the-fly search and near-instant indexed search.'
const SEARCH_TIMED_OUT_DEFAULT_TITLE = 'Search timed out'

interface SearchResultsListProps {
    isLightTheme: boolean
    location: H.Location
    user: GQL.IUser | null

    // Result list
    resultsOrError?: GQL.ISearchResults | ErrorLike
    uiLimit: number
    onShowMoreResultsClick: () => void

    // Expand all feature
    allExpanded: boolean
    onExpandAllResultsToggle: () => void

    // Saved queries
    showSavedQueryModal: boolean
    onSavedQueryModalClose: () => void
    onDidCreateSavedQuery: () => void
    onSaveQueryClick: () => void
    didSave: boolean
}

export class SearchResultsList extends React.PureComponent<SearchResultsListProps, {}> {
    public render(): React.ReactNode {
        let alert: {
            title: string
            description?: string | null
            proposedQueries?: GQL.ISearchQueryDescription[]
            errorBody?: React.ReactFragment
        } | null = null
        const searchTimeoutParameterEnabled = window.context.searchTimeoutParameterEnabled
        if (this.props.resultsOrError) {
            if (isErrorLike(this.props.resultsOrError)) {
                const error = this.props.resultsOrError
                if (error.message.includes('no query terms or regexp specified')) {
                    alert = { title: '', description: 'Enter terms to search...' }
                } else {
                    alert = { title: 'Something went wrong', description: upperFirst(error.message) }
                }
            } else {
                const results = this.props.resultsOrError
                if (results.alert) {
                    alert = results.alert
                } else if (
                    results.results.length === 0 &&
                    results.missing.length === 0 &&
                    results.cloning.length === 0
                ) {
                    const defaultTimeoutAlert = {
                        title: SEARCH_TIMED_OUT_DEFAULT_TITLE,
                        description: searchTimeoutParameterEnabled
                            ? "Try narrowing your query, or specifying a longer 'timeout:' in your query."
                            : 'Try narrowing your query.',
                    }
                    const longerTimeoutString = searchTimeoutParameterEnabled
                        ? "Specify a longer 'timeout:' in your query."
                        : ''
                    if (results.timedout.length > 0) {
                        if (window.context.sourcegraphDotComMode) {
                            alert = defaultTimeoutAlert
                        } else {
                            if (window.context.likelyDockerOnMac) {
                                alert = {
                                    title: SEARCH_TIMED_OUT_DEFAULT_TITLE,
                                    errorBody: this.renderSearchAlertTimeoutDetails([
                                        longerTimeoutString,
                                        DATA_CENTER_UPGRADE_STRING,
                                        'Use Docker Machine instead of Docker for Mac for better performance on macOS.',
                                    ]),
                                }
                            } else if (!window.context.likelyDockerOnMac && !window.context.isRunningDataCenter) {
                                alert = {
                                    title: SEARCH_TIMED_OUT_DEFAULT_TITLE,
                                    errorBody: this.renderSearchAlertTimeoutDetails([
                                        longerTimeoutString,
                                        DATA_CENTER_UPGRADE_STRING,
                                        'Run Sourcegraph on a server with more CPU and memory, or faster disk IO.',
                                    ]),
                                }
                            } else {
                                alert = defaultTimeoutAlert
                            }
                        }
                    } else {
                        alert = { title: 'No results' }
                    }
                }
            }
        }

        const parsedQuery = parseSearchURLQuery(this.props.location.search)

        return (
            <div className="search-results-list">
                {/* Saved Queries Form */}
                {this.props.showSavedQueryModal && (
                    <ModalContainer
                        onClose={this.props.onSavedQueryModalClose}
                        component={
                            <SavedQueryCreateForm
                                user={this.props.user}
                                values={{ query: parsedQuery ? parsedQuery.query : '' }}
                                onDidCancel={this.props.onSavedQueryModalClose}
                                onDidCreate={this.props.onDidCreateSavedQuery}
                            />
                        }
                    />
                )}

                {/* Loader */}
                {this.props.resultsOrError === undefined && <Loader className="icon-inline" />}

                {isSearchResults(this.props.resultsOrError) &&
                    (() => {
                        const results = this.props.resultsOrError
                        return (
                            <>
                                {/* Info Bar */}
                                <SearchResultsInfoBar
                                    user={this.props.user}
                                    results={results}
                                    allExpanded={this.props.allExpanded}
                                    didSave={this.props.didSave}
                                    onDidCreateSavedQuery={this.props.onDidCreateSavedQuery}
                                    onExpandAllResultsToggle={this.props.onExpandAllResultsToggle}
                                    onSaveQueryClick={this.props.onSaveQueryClick}
                                />

                                {/* Results */}
                                {results.results
                                    .slice(0, this.props.uiLimit)
                                    .map((result, i) => this.renderResult(i, result, i <= 15))}

                                {/* Show more button */}
                                {(results.limitHit || results.results.length > this.props.uiLimit) && (
                                    <button
                                        className="btn btn-link search-results-list__more"
                                        onClick={this.props.onShowMoreResultsClick}
                                    >
                                        Show more
                                    </button>
                                )}
                            </>
                        )
                    })()}
                {alert && (
                    <SearchAlert
                        className="search-results-list__alert"
                        title={alert.title}
                        description={alert.description || undefined}
                        proposedQueries={alert.proposedQueries}
                        location={this.props.location}
                        errorBody={alert.errorBody}
                    />
                )}
            </div>
        )
    }

    private renderSearchAlertTimeoutDetails(items: string[]): React.ReactFragment {
        return (
            <div className="search-alert__list">
                <p className="search-alert__list-header">Recommendations:</p>
                <ul className="search-alert__list-items">
                    {items.map(
                        (item, i) =>
                            item && (
                                <li key={i} className="search-alert__list-item">
                                    {item}
                                </li>
                            )
                    )}
                </ul>
            </div>
        )
    }

    private renderResult(key: number, result: GQL.SearchResult, expanded: boolean): React.ReactNode {
        switch (result.__typename) {
            case 'Repository':
                return <RepositorySearchResult key={key} result={result} onSelect={this.logEvent} />
            case 'FileMatch':
                return (
                    <FileMatch
                        key={key}
                        icon={result.lineMatches && result.lineMatches.length > 0 ? RepoIcon : DocumentIcon}
                        result={result}
                        onSelect={this.logEvent}
                        expanded={false}
                        showAllMatches={false}
                        isLightTheme={this.props.isLightTheme}
                        allExpanded={this.props.allExpanded}
                    />
                )
            case 'CommitSearchResult':
                return (
                    <CommitSearchResult
                        key={key}
                        location={this.props.location}
                        result={result}
                        onSelect={this.logEvent}
                        expanded={expanded}
                        allExpanded={this.props.allExpanded}
                    />
                )
        }
        return undefined
    }

    private logEvent = () => eventLogger.log('SearchResultClicked')
}
