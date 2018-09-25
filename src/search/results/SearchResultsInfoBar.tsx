import ArrowCollapseVerticalIcon from 'mdi-react/ArrowCollapseVerticalIcon'
import ArrowExpandVerticalIcon from 'mdi-react/ArrowExpandVerticalIcon'
import CalculatorIcon from 'mdi-react/CalculatorIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import CloudDownloadIcon from 'mdi-react/CloudDownloadIcon'
import DownloadIcon from 'mdi-react/DownloadIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import * as React from 'react'
import * as GQL from '../../backend/graphqlschema'
import { ServerBanner } from '../../marketing/ServerBanner'
import { showDotComMarketing } from '../../util/features'
import { pluralize } from '../../util/strings'

interface SearchResultsInfoBarProps {
    /** The logged-in user or null */
    user: GQL.IUser | null

    /** The loaded search results and metadata */
    results: GQL.ISearchResults
    onShowMoreResultsClick: () => void

    // Expand all feature
    allExpanded: boolean
    onExpandAllResultsToggle: () => void

    // Saved queries
    onDidCreateSavedQuery: () => void
    onSaveQueryClick: () => void
    didSave: boolean
}

/**
 * The info bar shown over the search results list that displays metadata
 * and a few actions like expand all and save query
 */
export const SearchResultsInfoBar: React.StatelessComponent<SearchResultsInfoBarProps> = props => (
    <div className="search-results-info-bar">
        {(props.results.timedout.length > 0 ||
            props.results.cloning.length > 0 ||
            props.results.results.length > 0 ||
            props.results.missing.length > 0) && (
            <small className="search-results-info-bar__row">
                <div className="search-results-info-bar__row-left">
                    {/* Time stats */}
                    {
                        <div className="search-results-info-bar__notice e2e-search-results-stats">
                            <span>
                                <CalculatorIcon className="icon-inline" /> {props.results.approximateResultCount}{' '}
                                {pluralize('result', props.results.resultCount)} in{' '}
                                {(props.results.elapsedMilliseconds / 1000).toFixed(2)} seconds
                                {props.results.indexUnavailable && ' (index unavailable)'}
                                {/* Nonbreaking space */}
                                {props.results.limitHit && String.fromCharCode(160)}
                            </span>
                            {/* Instantly accessible "show more button" */}
                            {props.results.limitHit && (
                                <button className="btn btn-link btn-sm p-0" onClick={props.onShowMoreResultsClick}>
                                    (show more)
                                </button>
                            )}
                        </div>
                    }
                    {/* Missing repos */}
                    {props.results.missing.length > 0 && (
                        <div
                            className="search-results-info-bar__notice"
                            data-tooltip={props.results.missing.map(repo => repo.name).join('\n')}
                        >
                            <span>
                                <MapSearchIcon className="icon-inline" /> {props.results.missing.length}{' '}
                                {pluralize('repository', props.results.missing.length, 'repositories')} not found
                            </span>
                        </div>
                    )}
                    {/* Timed out repos */}
                    {props.results.timedout.length > 0 && (
                        <div
                            className="search-results-info-bar__notice"
                            data-tooltip={props.results.timedout.map(repo => repo.name).join('\n')}
                        >
                            <span>
                                <TimerSandIcon className="icon-inline" /> {props.results.timedout.length}{' '}
                                {pluralize('repository', props.results.timedout.length, 'repositories')} timed out
                                (reload to try again, or specify a longer "timeout:" in your query)
                            </span>
                        </div>
                    )}
                    {/* Cloning repos */}
                    {props.results.cloning.length > 0 && (
                        <div
                            className="search-results-info-bar__notice"
                            data-tooltip={props.results.cloning.map(repo => repo.name).join('\n')}
                        >
                            <span>
                                <CloudDownloadIcon className="icon-inline" /> {props.results.cloning.length}{' '}
                                {pluralize('repository', props.results.cloning.length, 'repositories')} cloning (reload
                                to try again)
                            </span>
                        </div>
                    )}
                </div>
                <div className="search-results-info-bar__row-right">
                    {/* Expand all feature */}
                    {props.results.results.length > 0 && (
                        <button onClick={props.onExpandAllResultsToggle} className="btn btn-link">
                            {props.allExpanded ? (
                                <>
                                    <ArrowCollapseVerticalIcon className="icon-inline" /> Collapse all
                                </>
                            ) : (
                                <>
                                    <ArrowExpandVerticalIcon className="icon-inline" /> Expand all
                                </>
                            )}
                        </button>
                    )}
                    {/* Saved Queries */}
                    {props.user && (
                        <button onClick={props.onSaveQueryClick} className="btn btn-link" disabled={props.didSave}>
                            {props.didSave ? (
                                <>
                                    <CheckIcon className="icon-inline" /> Query saved
                                </>
                            ) : (
                                <>
                                    <DownloadIcon className="icon-inline" /> Save this search query
                                </>
                            )}
                        </button>
                    )}
                </div>
            </small>
        )}
        {!props.results.alert && showDotComMarketing && <ServerBanner />}
    </div>
)
