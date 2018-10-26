import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import FileIcon from 'mdi-react/FileIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { buildSearchURLQuery, parseSearchURLQuery } from '..'
import * as GQL from '../../backend/graphqlschema'
import { FileMatch } from '../../components/FileMatch'
import { ModalContainer } from '../../components/ModalContainer'
import { eventLogger } from '../../tracking/eventLogger'
import { ErrorLike, isErrorLike } from '../../util/errors'
import { RepositoryIcon } from '../../util/icons' // TODO: Switch to mdi icon
import { SavedQueryCreateForm } from '../saved-queries/SavedQueryCreateForm'
import { CommitSearchResult } from './CommitSearchResult'
import { RepositorySearchResult } from './RepositorySearchResult'
import { SearchResultsInfoBar } from './SearchResultsInfoBar'

interface SearchResultsListProps {
    isLightTheme: boolean
    location: H.Location
    authenticatedUser: GQL.IUser | null

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

export class SearchResultsListOld extends React.PureComponent<SearchResultsListProps, {}> {
    public render(): React.ReactNode {
        const parsedQuery = parseSearchURLQuery(this.props.location.search)

        return (
            <div className="search-results-list">
                {/* Saved Queries Form */}
                {this.props.showSavedQueryModal && (
                    <ModalContainer
                        onClose={this.props.onSavedQueryModalClose}
                        component={
                            <SavedQueryCreateForm
                                authenticatedUser={this.props.authenticatedUser}
                                values={{ query: parsedQuery ? parsedQuery.query : '' }}
                                onDidCancel={this.props.onSavedQueryModalClose}
                                onDidCreate={this.props.onDidCreateSavedQuery}
                            />
                        }
                    />
                )}

                {this.props.resultsOrError === undefined ? (
                    <div className="text-center">
                        <LoadingSpinner className="icon-inline" /> Loading
                    </div>
                ) : isErrorLike(this.props.resultsOrError) ? (
                    /* GraphQL, network, query syntax error */
                    <div className="alert alert-warning">
                        <AlertCircleIcon className="icon-inline" />
                        {upperFirst(this.props.resultsOrError.message)}
                    </div>
                ) : (
                    (() => {
                        const results = this.props.resultsOrError
                        return (
                            <>
                                {/* Info Bar */}
                                <SearchResultsInfoBar
                                    authenticatedUser={this.props.authenticatedUser}
                                    results={results}
                                    allExpanded={this.props.allExpanded}
                                    didSave={this.props.didSave}
                                    onDidCreateSavedQuery={this.props.onDidCreateSavedQuery}
                                    onExpandAllResultsToggle={this.props.onExpandAllResultsToggle}
                                    onSaveQueryClick={this.props.onSaveQueryClick}
                                    onShowMoreResultsClick={this.props.onShowMoreResultsClick}
                                />

                                {/* Results */}
                                {results.results
                                    .slice(0, this.props.uiLimit)
                                    .map((result, i) => this.renderResult(result, i <= 15))}

                                {/* Show more button */}
                                {(results.limitHit || results.results.length > this.props.uiLimit) && (
                                    <button
                                        className="btn btn-secondary btn-block"
                                        onClick={this.props.onShowMoreResultsClick}
                                    >
                                        Show more
                                    </button>
                                )}

                                {/* Server-provided help message */}
                                {results.alert ? (
                                    <div className="alert alert-info">
                                        <h3>
                                            <AlertCircleIcon className="icon-inline" /> {results.alert.title}
                                        </h3>
                                        <p>{results.alert.description}</p>
                                        {results.alert.proposedQueries && (
                                            <>
                                                <h4>Did you mean:</h4>
                                                <ul className="list-unstyled">
                                                    {results.alert.proposedQueries.map(proposedQuery => (
                                                        <li key={proposedQuery.query}>
                                                            <Link
                                                                className="btn btn-secondary btn-sm"
                                                                to={'/search?' + buildSearchURLQuery(proposedQuery)}
                                                            >
                                                                {proposedQuery.query || proposedQuery.description}
                                                            </Link>
                                                            {proposedQuery.query &&
                                                                proposedQuery.description &&
                                                                ` — ${proposedQuery.description}`}
                                                        </li>
                                                    ))}
                                                </ul>
                                            </>
                                        )}{' '}
                                    </div>
                                ) : (
                                    results.results.length === 0 &&
                                    (results.timedout.length > 0 ? (
                                        /* No results, but timeout hit */
                                        <div className="alert alert-warning">
                                            <h3>
                                                <TimerSandIcon className="icon-inline" /> Search timed out
                                            </h3>
                                            {this.renderRecommendations([
                                                <>
                                                    Try narrowing your query , or specifying a longer "timeout:" in your
                                                    query.
                                                </>,
                                                /* If running on non-cluster, give some smart advice */
                                                ...(!window.context.sourcegraphDotComMode &&
                                                !window.context.isClusterDeployment
                                                    ? [
                                                          <>
                                                              Upgrade to Sourcegraph Enterprise for a highly scalable
                                                              Kubernetes cluster deployment option.
                                                          </>,
                                                          window.context.likelyDockerOnMac
                                                              ? 'Use Docker Machine instead of Docker for Mac for better performance on macOS'
                                                              : 'Run Sourcegraph on a server with more CPU and memory, or faster disk IO',
                                                      ]
                                                    : []),
                                            ])}
                                        </div>
                                    ) : (
                                        <>
                                            <div className="alert alert-info d-flex">
                                                <h3 className="m-0">
                                                    <SearchIcon className="icon-inline" /> No results
                                                </h3>
                                            </div>
                                        </>
                                    ))
                                )}
                            </>
                        )
                    })()
                )}
                <div className="pb-4" />
                {this.props.resultsOrError !== undefined && (
                    <Link className="mb-2" to="/help/user/search">
                        Not seeing expected results?
                    </Link>
                )}
            </div>
        )
    }

    /**
     * Renders the given recommendations in a list if multiple, otherwise returns the first one or undefined
     */
    private renderRecommendations(recommendations: React.ReactNode[]): React.ReactNode {
        if (recommendations.length <= 1) {
            return recommendations[0]
        }
        return (
            <>
                <h4>Recommendations:</h4>
                <ul>{recommendations.map((recommendation, i) => <li key={i}>{recommendation}</li>)}</ul>
            </>
        )
    }

    private renderResult(result: GQL.SearchResult, expanded: boolean): JSX.Element | undefined {
        switch (result.__typename) {
            case 'Repository':
                return <RepositorySearchResult key={'repo:' + result.id} result={result} onSelect={this.logEvent} />
            case 'FileMatch':
                return (
                    <FileMatch
                        key={'file:' + result.file.url}
                        icon={result.lineMatches && result.lineMatches.length > 0 ? RepositoryIcon : FileIcon}
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
                        key={'commit:' + result.commit.id}
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
