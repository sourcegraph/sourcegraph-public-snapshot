import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CalculatorIcon from 'mdi-react/CalculatorIcon'
import CloudDownloadIcon from 'mdi-react/CloudDownloadIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React from 'react'

import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

/** Search result statistics for GraphQL searches */
export const SearchResultsStats: React.FunctionComponent<{
    results: GQL.ISearchResults
    onShowMoreResultsClick?: () => void
}> = ({ results, onShowMoreResultsClick }) => {
    const excludeForksFilter = results.dynamicFilters.find(filter => filter.value === 'fork:yes')
    const excludedForksCount = excludeForksFilter?.count || 0
    const excludeArchivedFilter = results.dynamicFilters.find(filter => filter.value === 'archived:yes')
    const excludedArchivedCount = excludeArchivedFilter?.count || 0

    return (
        <>
            <div className="search-results-info-bar__notice test-search-results-stats">
                <small>
                    <CalculatorIcon className="icon-inline" /> {results.approximateResultCount}{' '}
                    {pluralize('result', results.matchCount)} in {(results.elapsedMilliseconds / 1000).toFixed(2)}{' '}
                    seconds
                    {results.indexUnavailable && ' (index unavailable)'}
                    {results.limitHit && String.fromCharCode(160)}
                </small>

                {results.limitHit && onShowMoreResultsClick && (
                    <button type="button" className="btn btn-link btn-sm p-0" onClick={onShowMoreResultsClick}>
                        (show more)
                    </button>
                )}
            </div>

            {excludedForksCount > 0 && (
                <div className="search-results-info-bar__notice" data-tooltip="add fork:yes to include forks">
                    <small>
                        <AlertCircleIcon className="icon-inline" /> {excludedForksCount} forked{' '}
                        {pluralize('repository', excludedForksCount, 'repositories')} excluded
                    </small>
                </div>
            )}

            {excludedArchivedCount > 0 && (
                <div className="search-results-info-bar__notice" data-tooltip="add archived:yes to include archives">
                    <small>
                        <AlertCircleIcon className="icon-inline" /> {excludedArchivedCount} archived{' '}
                        {pluralize('repository', excludedArchivedCount, 'repositories')} excluded
                    </small>
                </div>
            )}

            {results.missing.length > 0 && (
                <div
                    className="search-results-info-bar__notice"
                    data-tooltip={results.missing.map(repo => repo.name).join('\n')}
                >
                    <small>
                        <MapSearchIcon className="icon-inline" /> {results.missing.length}{' '}
                        {pluralize('repository', results.missing.length, 'repositories')} not found
                    </small>
                </div>
            )}

            {results.timedout.length > 0 && (
                <div
                    className="search-results-info-bar__notice"
                    data-tooltip={results.timedout.map(repo => repo.name).join('\n')}
                >
                    <small>
                        <TimerSandIcon className="icon-inline" /> {results.timedout.length}{' '}
                        {pluralize('repository', results.timedout.length, 'repositories')} timed out (reload to try
                        again, or specify a longer "timeout:" in your query)
                    </small>
                </div>
            )}

            {results.cloning.length > 0 && (
                <div
                    className="search-results-info-bar__notice"
                    data-tooltip={results.cloning.map(repo => repo.name).join('\n')}
                >
                    <small>
                        <CloudDownloadIcon className="icon-inline" /> {results.cloning.length}{' '}
                        {pluralize('repository', results.cloning.length, 'repositories')} cloning (reload to try again)
                    </small>
                </div>
            )}
        </>
    )
}
