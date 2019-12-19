import H from 'history'
import ArrowCollapseVerticalIcon from 'mdi-react/ArrowCollapseVerticalIcon'
import ArrowExpandVerticalIcon from 'mdi-react/ArrowExpandVerticalIcon'
import CalculatorIcon from 'mdi-react/CalculatorIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import CloudDownloadIcon from 'mdi-react/CloudDownloadIcon'
import DownloadIcon from 'mdi-react/DownloadIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import FormatQuoteOpenIcon from 'mdi-react/FormatQuoteOpenIcon'
import * as React from 'react'
import { ContributableMenu } from '../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { pluralize } from '../../../../shared/src/util/strings'
import { WebActionsNavItems as ActionsNavItems } from '../../components/shared'
import { ServerBanner, ServerBannerNoRepo } from '../../marketing/ServerBanner'
import { PerformanceWarningAlert } from '../../site/PerformanceWarningAlert'
import { PatternTypeProps } from '..'

interface SearchResultsInfoBarProps
    extends ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip'>,
        TelemetryProps,
        PatternTypeProps {
    /** The currently authenticated user or null */
    authenticatedUser: GQL.IUser | null

    /** The loaded search results and metadata */
    query?: string
    results: GQL.ISearchResults
    onShowMoreResultsClick: () => void

    // Expand all feature
    allExpanded: boolean
    onExpandAllResultsToggle: () => void

    showDotComMarketing: boolean
    // Saved queries
    onDidCreateSavedQuery: () => void
    onSaveQueryClick: () => void
    didSave: boolean

    displayPerformanceWarning: boolean

    // Whether the search query contains a repo: field.
    hasRepoishField: boolean

    location: H.Location
}

/**
 * A notice for when the user is searching literally and has quotes in thier
 * query, in which case it is possible that they think their query `"foobar"`
 * will be searching literally for `foobar` (without quotes). This notice
 * informs them that this may be the case to avoid confusion.
 */
const QuotesInterpretedLiterallyNotice: React.FunctionComponent<SearchResultsInfoBarProps> = props =>
    props.patternType === GQL.SearchPatternType.literal && props.query && props.query.includes('"') ? (
        <div
            className={`search-results-info-bar__notice${
                props.results.results.length === 0 ? ' search-results-info-bar__notice--no-results' : ''
            }`}
            data-tooltip="Your search query is interpreted literally, including the quotes. Use the .* toggle to switch between literal and regular expression search."
        >
            <span>
                <FormatQuoteOpenIcon className="icon-inline" />
                Searching literally <strong>(including quotes)</strong>
            </span>
        </div>
    ) : null

/**
 * The info bar shown over the search results list that displays metadata
 * and a few actions like expand all and save query
 */
export const SearchResultsInfoBar: React.FunctionComponent<SearchResultsInfoBarProps> = props => (
    <div className="search-results-info-bar" data-testid="results-info-bar">
        {/*
            If there were no results, still show the "quotes are interpreted literally"
            notice as this is the most common case where a user will make this mistake.
        */}
        {props.results.results.length === 0 && (
            <small className="search-results-info-bar__row">
                <div className="search-results-info-bar__row-left">
                    <QuotesInterpretedLiterallyNotice {...props} />
                </div>
                <ul className="search-results-info-bar__row-right nav align-items-center justify-content-end"></ul>
            </small>
        )}
        {(props.results.timedout.length > 0 ||
            props.results.cloning.length > 0 ||
            props.results.results.length > 0 ||
            props.results.missing.length > 0) && (
            <small className="search-results-info-bar__row">
                <div className="search-results-info-bar__row-left">
                    {/* Time stats */}
                    <div className="search-results-info-bar__notice e2e-search-results-stats">
                        <span>
                            <CalculatorIcon className="icon-inline" /> {props.results.approximateResultCount}{' '}
                            {pluralize('result', props.results.matchCount)} in{' '}
                            {(props.results.elapsedMilliseconds / 1000).toFixed(2)} seconds
                            {props.results.indexUnavailable && ' (index unavailable)'}
                            {/* Nonbreaking space */}
                            {props.results.limitHit && String.fromCharCode(160)}
                        </span>
                        {/* Instantly accessible "show more button" */}
                        {props.results.limitHit && (
                            <button
                                type="button"
                                className="btn btn-link btn-sm p-0"
                                onClick={props.onShowMoreResultsClick}
                            >
                                (show more)
                            </button>
                        )}
                    </div>

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
                    <QuotesInterpretedLiterallyNotice {...props} />
                </div>
                <ul className="search-results-info-bar__row-right nav align-items-center justify-content-end">
                    <ActionsNavItems
                        {...props}
                        extraContext={{ searchQuery: props.query }}
                        menu={ContributableMenu.SearchResultsToolbar}
                        wrapInList={false}
                        showLoadingSpinnerDuringExecution={true}
                        actionItemClass="btn btn-link nav-link text-decoration-none"
                    />
                    {/* Expand all feature */}
                    {props.results.results.length > 0 && (
                        <li className="nav-item">
                            <button
                                type="button"
                                onClick={props.onExpandAllResultsToggle}
                                className="btn btn-link nav-link text-decoration-none"
                                data-tooltip={`${props.allExpanded ? 'Hide' : 'Show'} more matches on all results`}
                            >
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
                        </li>
                    )}
                    {/* Saved Queries */}
                    {props.authenticatedUser && (
                        <li className="nav-item">
                            <button
                                type="button"
                                onClick={props.onSaveQueryClick}
                                className="btn btn-link nav-link text-decoration-none"
                                disabled={props.didSave}
                            >
                                {props.didSave ? (
                                    <>
                                        <CheckIcon className="icon-inline" /> Query saved
                                    </>
                                ) : (
                                    <>
                                        <DownloadIcon className="icon-inline e2e-save-search-link" /> Save this search
                                        query
                                    </>
                                )}
                            </button>
                        </li>
                    )}
                </ul>
            </small>
        )}
        {!props.results.alert &&
            props.showDotComMarketing &&
            (props.hasRepoishField ? <ServerBanner /> : <ServerBannerNoRepo />)}
        {!props.results.alert && props.displayPerformanceWarning && <PerformanceWarningAlert />}
    </div>
)
