import classNames from 'classnames'
import * as H from 'history'
import ArrowCollapseUpIcon from 'mdi-react/ArrowCollapseUpIcon'
import ArrowCollapseVerticalIcon from 'mdi-react/ArrowCollapseVerticalIcon'
import ArrowExpandDownIcon from 'mdi-react/ArrowExpandDownIcon'
import ArrowExpandVerticalIcon from 'mdi-react/ArrowExpandVerticalIcon'
import DownloadIcon from 'mdi-react/DownloadIcon'
import FormatQuoteOpenIcon from 'mdi-react/FormatQuoteOpenIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuIcon from 'mdi-react/MenuIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { useMemo, useState } from 'react'

import { ContributableMenu } from '@sourcegraph/shared/src/api/protocol'
import { ButtonLink } from '@sourcegraph/shared/src/components/LinkOrButton'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/validate'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

import { PatternTypeProps, CaseSensitivityProps } from '..'
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
        Pick<CaseSensitivityProps, 'caseSensitive'>,
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

    onShowFiltersChanged?: (show: boolean) => void
}

/**
 * A notice for when the user is searching literally and has quotes in their
 * query, in which case it is possible that they think their query `"foobar"`
 * will be searching literally for `foobar` (without quotes). This notice
 * informs them that this may be the case to avoid confusion.
 */
const QuotesInterpretedLiterallyNotice: React.FunctionComponent<SearchResultsInfoBarProps> = props =>
    props.patternType === SearchPatternType.literal && props.query && props.query.includes('"') ? (
        <small
            className="search-results-info-bar__notice"
            data-tooltip="Your search query is interpreted literally, including the quotes. Use the .* toggle to switch between literal and regular expression search."
        >
            <span>
                <FormatQuoteOpenIcon className="icon-inline" />
                Searching literally <strong>(including quotes)</strong>
            </span>
        </small>
    ) : null

/**
 * The info bar shown over the search results list that displays metadata
 * and a few actions like expand all and save query
 */
export const SearchResultsInfoBar: React.FunctionComponent<SearchResultsInfoBarProps> = props => {
    const [isRedesignEnabled] = useRedesignToggle()
    const buttonClass = isRedesignEnabled ? 'btn-outline-secondary mr-2' : 'btn-link'

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
                    className={classNames('btn btn-sm nav-link text-decoration-none', buttonClass)}
                >
                    <CodeMonitoringLogo className="icon-inline mr-1" />
                    Monitor
                </ButtonLink>
            </li>
        )
    }, [
        buttonClass,
        props.enableCodeMonitoring,
        props.query,
        props.authenticatedUser,
        props.location.search,
        props.patternType,
    ])

    const saveSearchButton = useMemo(() => {
        if (props.showSavedQueryButton === false || !props.authenticatedUser) {
            return null
        }

        return (
            <li className="nav-item">
                <button
                    type="button"
                    onClick={props.onSaveQueryClick}
                    className={classNames(
                        'btn btn-sm nav-link text-decoration-none test-save-search-link',
                        buttonClass
                    )}
                >
                    <DownloadIcon className="icon-inline mr-1" />
                    Save search
                </button>
            </li>
        )
    }, [buttonClass, props.authenticatedUser, props.onSaveQueryClick, props.showSavedQueryButton])

    const extraContext = useMemo(
        () => ({
            searchQuery: props.query || null,
            patternType: props.patternType,
            caseSensitive: props.caseSensitive,
        }),
        [props.query, props.patternType, props.caseSensitive]
    )

    const [showFilters, setShowFilters] = useState(false)
    const onShowFiltersClicked = (): void => {
        const newShowFilters = !showFilters
        setShowFilters(newShowFilters)
        props.onShowFiltersChanged?.(newShowFilters)
    }

    return (
        <div className={classNames(props.className, 'search-results-info-bar')} data-testid="results-info-bar">
            <div className="search-results-info-bar__row">
                {isRedesignEnabled && (
                    <button
                        type="button"
                        className={classNames(
                            'btn btn-sm btn-outline-secondary d-flex d-lg-none',
                            showFilters && 'active'
                        )}
                        aria-pressed={showFilters}
                        onClick={onShowFiltersClicked}
                    >
                        <MenuIcon className="icon-inline mr-1" />
                        Filters
                        {showFilters ? (
                            <MenuUpIcon className="icon-inline" />
                        ) : (
                            <MenuDownIcon className="icon-inline" />
                        )}
                    </button>
                )}
                {props.stats}
                <QuotesInterpretedLiterallyNotice {...props} />

                <div className="search-results-info-bar__expander" />

                <ul className={classNames('nav align-items-center', !isRedesignEnabled && 'justify-content-end')}>
                    <ActionsNavItems
                        {...props}
                        extraContext={extraContext}
                        menu={ContributableMenu.SearchResultsToolbar}
                        wrapInList={false}
                        showLoadingSpinnerDuringExecution={true}
                        actionItemClass={classNames('btn nav-link text-decoration-none btn-sm', buttonClass)}
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
                                    className={classNames('btn btn-sm nav-link text-decoration-none', buttonClass)}
                                    data-tooltip={`${props.allExpanded ? 'Hide' : 'Show'} more matches on all results`}
                                >
                                    {props.allExpanded ? (
                                        isRedesignEnabled ? (
                                            <ArrowCollapseUpIcon className="icon-inline mr-0" />
                                        ) : (
                                            <ArrowCollapseVerticalIcon className="icon-inline mr-0" />
                                        )
                                    ) : isRedesignEnabled ? (
                                        <ArrowExpandDownIcon className="icon-inline mr-0" />
                                    ) : (
                                        <ArrowExpandVerticalIcon className="icon-inline mr-0" />
                                    )}
                                </button>
                            </li>
                        </>
                    )}
                </ul>
            </div>
        </div>
    )
}
