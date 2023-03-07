import React, { useMemo, useState } from 'react'

import { mdiChevronDoubleDown, mdiChevronDoubleUp } from '@mdi/js'
import classNames from 'classnames'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { CaseSensitivityProps, SearchPatternTypeProps } from '@sourcegraph/shared/src/search'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon, Label } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'

import {
    CreateAction,
    getBatchChangeCreateAction,
    getCodeMonitoringCreateAction,
    getInsightsCreateAction,
    getSearchContextCreateAction,
} from './createActions'
import { SearchActionsMenu } from './SearchActionsMenu'

import styles from './SearchResultsInfoBar.module.scss'

export interface SearchResultsInfoBarProps
    extends TelemetryProps,
        PlatformContextProps<'settings' | 'sourcegraphURL'>,
        SearchPatternTypeProps,
        Pick<CaseSensitivityProps, 'caseSensitive'> {
    /** The currently authenticated user or null */
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'displayName' | 'emails'> | null

    enableCodeInsights?: boolean
    enableCodeMonitoring: boolean

    /** The search query and results */
    query?: string
    results?: AggregateStreamingSearchResults

    batchChangesEnabled?: boolean
    /** Whether running batch changes server-side is enabled */
    batchChangesExecutionEnabled?: boolean

    // Expand all feature
    allExpanded: boolean
    onExpandAllResultsToggle: () => void

    // Saved queries
    onSaveQueryClick: () => void

    className?: string

    stats: JSX.Element

    onShowMobileFiltersChanged?: (show: boolean) => void

    sidebarCollapsed: boolean
    setSidebarCollapsed: (collapsed: boolean) => void

    isSourcegraphDotCom: boolean

    isRankingEnabled: boolean
    setRankingEnabled: (enabled: boolean) => void
}

/**
 * The info bar shown over the search results list that displays metadata
 * and a few actions like expand all and save query
 */
export const SearchResultsInfoBar: React.FunctionComponent<
    React.PropsWithChildren<SearchResultsInfoBarProps>
> = props => {
    const globalTypeFilter = useMemo(
        () => (props.query ? findFilter(props.query, 'type', FilterKind.Global)?.value?.value : undefined),
        [props.query]
    )

    const canCreateMonitorFromQuery = useMemo(
        () => globalTypeFilter === 'diff' || globalTypeFilter === 'commit',
        [globalTypeFilter]
    )

    const canCreateBatchChangeFromQuery = useMemo(
        () => globalTypeFilter !== 'diff' && globalTypeFilter !== 'commit',
        [globalTypeFilter]
    )

    // When adding a new create action check and update the $collapse-breakpoint in CreateActions.module.scss.
    // The collapse breakpoint indicates at which window size we hide the buttons and show the collapsed menu instead.
    const createActions = useMemo(
        () =>
            [
                getBatchChangeCreateAction(
                    props.query,
                    props.patternType,
                    Boolean(
                        props.batchChangesEnabled &&
                            props.batchChangesExecutionEnabled &&
                            props.authenticatedUser &&
                            canCreateBatchChangeFromQuery
                    )
                ),
                getSearchContextCreateAction(props.query, props.authenticatedUser),
                getInsightsCreateAction(props.query, props.patternType, window.context?.codeInsightsEnabled),
            ].filter((button): button is CreateAction => button !== null),
        [
            props.authenticatedUser,
            props.patternType,
            props.query,
            props.batchChangesEnabled,
            props.batchChangesExecutionEnabled,
            canCreateBatchChangeFromQuery,
        ]
    )

    // The create code monitor action is separated from the rest of the actions, because we use the
    // <ExperimentalActionButton /> component instead of a regular (button) link, and it has a tour attached.
    const createCodeMonitorAction = useMemo(
        () => getCodeMonitoringCreateAction(props.query, props.patternType, props.enableCodeMonitoring),
        [props.enableCodeMonitoring, props.patternType, props.query]
    )

    // Show/hide mobile filters menu
    const [showMobileFilters, setShowMobileFilters] = useState(false)
    const onShowMobileFiltersClicked = (): void => {
        const newShowFilters = !showMobileFilters
        setShowMobileFilters(newShowFilters)
        props.onShowMobileFiltersChanged?.(newShowFilters)
    }

    // Show/hide ranking toggle
    const [rankingEnabled] = useFeatureFlag('search-ranking')

    return (
        <aside
            role="region"
            aria-label="Search results information"
            className={classNames(props.className, styles.searchResultsInfoBar)}
            data-testid="results-info-bar"
        >
            <div className={styles.row}>
                {props.stats}

                <div className={styles.expander} />

                {rankingEnabled && (
                    <Label className={styles.toggle}>
                        Intelligent ranking{' '}
                        <Toggle
                            value={props.isRankingEnabled}
                            onToggle={() => props.setRankingEnabled(!props.isRankingEnabled)}
                            title="Enable Ranking"
                            className="mr-2"
                        />
                    </Label>
                )}
                <ul className="nav align-items-center">
                    <SearchActionsMenu
                        query={props.query}
                        patternType={props.patternType}
                        sourcegraphURL={props.platformContext.sourcegraphURL}
                        authenticatedUser={props.authenticatedUser}
                        createActions={createActions}
                        createCodeMonitorAction={createCodeMonitorAction}
                        canCreateMonitor={canCreateMonitorFromQuery}
                        results={props.results}
                        allExpanded={props.allExpanded}
                        onExpandAllResultsToggle={props.onExpandAllResultsToggle}
                        onSaveQueryClick={props.onSaveQueryClick}
                        telemetryService={props.telemetryService}
                    />
                </ul>

                <Button
                    className={classNames(
                        'd-flex align-items-center d-lg-none',
                        styles.filtersButton,
                        showMobileFilters && 'active'
                    )}
                    aria-pressed={showMobileFilters}
                    onClick={onShowMobileFiltersClicked}
                    outline={true}
                    variant="secondary"
                    size="sm"
                    aria-label={`${showMobileFilters ? 'Hide' : 'Show'} filters`}
                >
                    Filters
                    <Icon
                        aria-hidden={true}
                        className="ml-2"
                        svgPath={showMobileFilters ? mdiChevronDoubleUp : mdiChevronDoubleDown}
                    />
                </Button>

                {props.sidebarCollapsed && (
                    <Button
                        className={classNames('align-items-center d-none d-lg-flex', styles.filtersButton)}
                        onClick={() => props.setSidebarCollapsed(false)}
                        outline={true}
                        variant="secondary"
                        size="sm"
                        aria-label="Show filters sidebar"
                    >
                        Filters
                        <Icon aria-hidden={true} className="ml-2" svgPath={mdiChevronDoubleDown} />
                    </Button>
                )}
            </div>
        </aside>
    )
}
