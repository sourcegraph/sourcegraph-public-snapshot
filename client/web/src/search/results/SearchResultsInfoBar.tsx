import classNames from 'classnames'
import * as H from 'history'
import ArrowCollapseVerticalIcon from 'mdi-react/ArrowCollapseVerticalIcon'
import ArrowExpandVerticalIcon from 'mdi-react/ArrowExpandVerticalIcon'
import DownloadIcon from 'mdi-react/DownloadIcon'
import FormatQuoteOpenIcon from 'mdi-react/FormatQuoteOpenIcon'
import React, { useMemo } from 'react'

import { ContributableMenu } from '@sourcegraph/shared/src/api/protocol'
import { ButtonLink } from '@sourcegraph/shared/src/components/LinkOrButton'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/validate'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { PatternTypeProps } from '..'
import { AuthenticatedUser } from '../../auth'
import { CodeMonitoringProps } from '../../code-monitoring'
import { CodeMonitoringLogo } from '../../code-monitoring/CodeMonitoringLogo'
import { WebActionsNavItems as ActionsNavItems } from '../../components/shared'
import { SearchPatternType } from '../../graphql-operations'

export interface SearchResultsInfoBarProps
    extends ExtensionsControllerProps<'executeCommand' | 'extHostAPI'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        TelemetryProps,
        Pick<PatternTypeProps, 'patternType'>,
        CodeMonitoringProps {
    history: H.History
    /** The currently authenticated user or null */
    authenticatedUser: Pick<AuthenticatedUser, 'id'> | null

    /** The search query and if any results were found */
    query?: string
    resultsFound: boolean

    // Expand all feature
    allExpanded: boolean
    onExpandAllResultsToggle: () => void

    // Saved queries
    showSavedQueryButton?: boolean
    onSaveQueryClick: () => void

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
export const SearchResultsInfoBar: React.FunctionComponent<SearchResultsInfoBarProps> = props => {
    const createCodeMonitorButton = useMemo(() => {
        if (!props.enableCodeMonitoring || !props.query || !props.authenticatedUser) {
            return null
        }
        const globalTypeFilterInQuery = findFilter(props.query, 'type', FilterKind.Global)
        const globalTypeFilterValue = globalTypeFilterInQuery?.value ? globalTypeFilterInQuery.value.value : undefined
        const canCreateMonitorFromQuery = globalTypeFilterValue === 'diff' || globalTypeFilterValue === 'commit'
        const searchParameters = new URLSearchParams(props.location.search)
        searchParameters.set('trigger-query', `${props.query} patterntype:${props.patternType}`)
        const toURL = `/code-monitoring/new?${searchParameters.toString()}`
        return (
            <li
                className="nav-item"
                data-tooltip={
                    !canCreateMonitorFromQuery
                        ? 'Code monitors only support type:diff or type:commit searches.'
                        : undefined
                }
            >
                <ButtonLink
                    disabled={!canCreateMonitorFromQuery}
                    to={toURL}
                    className="btn btn-sm btn-link nav-link text-decoration-none"
                >
                    <CodeMonitoringLogo className="icon-inline mr-1" />
                    Monitor
                </ButtonLink>
            </li>
        )
    }, [props.enableCodeMonitoring, props.query, props.authenticatedUser, props.location.search, props.patternType])

    const saveSearchButton = useMemo(() => {
        if (props.showSavedQueryButton === false || !props.authenticatedUser) {
            return null
        }

        return (
            <li className="nav-item">
                <button
                    type="button"
                    onClick={props.onSaveQueryClick}
                    className="btn btn-sm btn-link nav-link text-decoration-none test-save-search-link"
                >
                    <DownloadIcon className="icon-inline mr-1" />
                    Save search
                </button>
            </li>
        )
    }, [props.authenticatedUser, props.onSaveQueryClick, props.showSavedQueryButton])

    const extraContext = useMemo(() => ({ searchQuery: props.query || null }), [props.query])

    return (
        <div className={classNames(props.className, 'search-results-info-bar')} data-testid="results-info-bar">
            <small className="search-results-info-bar__row">
                {props.stats}
                <QuotesInterpretedLiterallyNotice {...props} />

                <ul className="nav align-items-center justify-content-end">
                    <ActionsNavItems
                        {...props}
                        extraContext={extraContext}
                        menu={ContributableMenu.SearchResultsToolbar}
                        wrapInList={false}
                        showLoadingSpinnerDuringExecution={true}
                        actionItemClass="btn btn-sm btn-link nav-link text-decoration-none"
                    />

                    {(createCodeMonitorButton || saveSearchButton) && (
                        <li className="search-results-info-bar__divider" aria-hidden="true" />
                    )}
                    {createCodeMonitorButton}
                    {saveSearchButton}

                    {props.resultsFound && (
                        <>
                            <li className="search-results-info-bar__divider" aria-hidden="true" />
                            <li className="nav-item">
                                <button
                                    type="button"
                                    onClick={props.onExpandAllResultsToggle}
                                    className="btn btn-sm btn-link nav-link text-decoration-none"
                                    data-tooltip={`${props.allExpanded ? 'Hide' : 'Show'} more matches on all results`}
                                >
                                    {props.allExpanded ? (
                                        <ArrowCollapseVerticalIcon className="icon-inline" />
                                    ) : (
                                        <ArrowExpandVerticalIcon className="icon-inline" />
                                    )}
                                </button>
                            </li>
                        </>
                    )}
                </ul>
            </small>
        </div>
    )
}
