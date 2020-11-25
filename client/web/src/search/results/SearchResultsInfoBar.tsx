import * as H from 'history'
import * as React from 'react'
import ArrowCollapseVerticalIcon from 'mdi-react/ArrowCollapseVerticalIcon'
import ArrowExpandVerticalIcon from 'mdi-react/ArrowExpandVerticalIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import classNames from 'classnames'
import DownloadIcon from 'mdi-react/DownloadIcon'
import FormatQuoteOpenIcon from 'mdi-react/FormatQuoteOpenIcon'
import { AuthenticatedUser } from '../../auth'
import { ContributableMenu } from '../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { PatternTypeProps } from '..'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SearchPatternType } from '../../graphql-operations'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { WebActionsNavItems as ActionsNavItems } from '../../components/shared'

interface SearchResultsInfoBarProps
    extends ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        TelemetryProps,
        PatternTypeProps {
    /** The currently authenticated user or null */
    authenticatedUser: AuthenticatedUser | null

    /** The search query and if any results were found */
    query?: string
    resultsFound: boolean

    // Expand all feature
    allExpanded: boolean
    onExpandAllResultsToggle: () => void

    // Saved queries
    showSavedQueryButton?: boolean
    onDidCreateSavedQuery: () => void
    onSaveQueryClick: () => void
    didSave: boolean

    location: H.Location

    className?: string

    stats: JSX.Element
}

/**
 * A notice for when the user is searching literally and has quotes in their
 * query, in which case it is possible that they think their query `"foobar"`
 * will be searching literally for `foobar` (without quotes). This notice
 * informs them that this may be the case to avoid confusion.
 */
const QuotesInterpretedLiterallyNotice: React.FunctionComponent<SearchResultsInfoBarProps> = props =>
    props.patternType === SearchPatternType.literal && props.query && props.query.includes('"') ? (
        <div
            className="search-results-info-bar__notice"
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
    <div className={classNames(props.className, 'search-results-info-bar')} data-testid="results-info-bar">
        <small className="search-results-info-bar__row">
            <div className="search-results-info-bar__row-left">
                {props.stats}
                <QuotesInterpretedLiterallyNotice {...props} />
            </div>

            <ul className="search-results-info-bar__row-right nav align-items-center justify-content-end">
                <ActionsNavItems
                    {...props}
                    extraContext={{ searchQuery: props.query || null }}
                    menu={ContributableMenu.SearchResultsToolbar}
                    wrapInList={false}
                    showLoadingSpinnerDuringExecution={true}
                    actionItemClass="btn btn-link nav-link text-decoration-none"
                />

                {props.resultsFound && (
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

                {props.showSavedQueryButton !== false && props.authenticatedUser && (
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
                                    <DownloadIcon className="icon-inline test-save-search-link" /> Save this search
                                    query
                                </>
                            )}
                        </button>
                    </li>
                )}
            </ul>
        </small>
    </div>
)
