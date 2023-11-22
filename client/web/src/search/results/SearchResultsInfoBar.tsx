import React, { useMemo, useState } from 'react'

import { mdiChevronDoubleDown, mdiChevronDoubleUp, mdiThumbUp, mdiThumbDown, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'
import { useLocation } from 'react-router-dom'

import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { CaseSensitivityProps, SearchPatternTypeProps } from '@sourcegraph/shared/src/search'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import type { AggregateStreamingSearchResults, StreamSearchOptions } from '@sourcegraph/shared/src/search/stream'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon, Alert, useSessionStorage, Link, Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { canWriteBatchChanges, NO_ACCESS_BATCH_CHANGES_WRITE, NO_ACCESS_SOURCEGRAPH_COM } from '../../batches/utils'
import { eventLogger } from '../../tracking/eventLogger'
import { DOTCOM_URL } from '../../tracking/util'

import {
    type CreateAction,
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
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'displayName' | 'emails' | 'permissions'> | null

    enableCodeInsights?: boolean
    enableCodeMonitoring: boolean

    /** The search query and results */
    query?: string
    options: StreamSearchOptions
    results?: AggregateStreamingSearchResults

    batchChangesEnabled?: boolean
    /** Whether running batch changes server-side is enabled */
    batchChangesExecutionEnabled?: boolean

    // Expand all feature
    allExpanded: boolean
    onExpandAllResultsToggle: () => void

    // Saved queries
    onSaveQueryClick: () => void

    // Download CSV of search results
    onExportCsvClick: () => void

    className?: string

    stats: JSX.Element

    onShowMobileFiltersChanged?: (show: boolean) => void

    sidebarCollapsed: boolean
    setSidebarCollapsed: (collapsed: boolean) => void

    isSourcegraphDotCom: boolean
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

    const canCreateBatchChanges: boolean | string = useMemo(() => {
        if (globalTypeFilter === 'diff' || globalTypeFilter === 'commit') {
            return 'Batch changes cannot be created from searches with type:diff or type:commit'
        }
        if (props.isSourcegraphDotCom) {
            return NO_ACCESS_SOURCEGRAPH_COM
        }
        if (!props.batchChangesEnabled || !props.batchChangesExecutionEnabled) {
            return false
        }
        if (!canWriteBatchChanges(props.authenticatedUser)) {
            return NO_ACCESS_BATCH_CHANGES_WRITE
        }
        return true
    }, [
        globalTypeFilter,
        props.isSourcegraphDotCom,
        props.batchChangesEnabled,
        props.batchChangesExecutionEnabled,
        props.authenticatedUser,
    ])

    // When adding a new create action check and update the $collapse-breakpoint in CreateActions.module.scss.
    // The collapse breakpoint indicates at which window size we hide the buttons and show the collapsed menu instead.
    const createActions = useMemo(
        () =>
            [
                getBatchChangeCreateAction(props.query, props.patternType, canCreateBatchChanges),
                getSearchContextCreateAction(props.query, props.authenticatedUser),
                getInsightsCreateAction(props.query, props.patternType, window.context?.codeInsightsEnabled),
            ].filter((button): button is CreateAction => button !== null),
        [props.authenticatedUser, props.patternType, props.query, canCreateBatchChanges]
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

    const location = useLocation()
    const dotcomHost = DOTCOM_URL.href
    const isPrivateInstance = window.location.host !== dotcomHost
    const refFromCodySearch = new URLSearchParams(location.search).get('ref') === 'cody-search'
    const [codySearchInputString] = useSessionStorage<string>('cody-search-input', '')
    const codySearchInput: { input?: string; translatedQuery?: string } = JSON.parse(codySearchInputString || '{}')
    const [codyFeedback, setCodyFeedback] = useState<null | boolean>(null)

    const collectCodyFeedback = (positive: boolean): void => {
        if (codyFeedback !== null) {
            return
        }

        eventLogger.log(
            'web:codySearch:feedbackSubmitted',
            !isPrivateInstance ? { ...codySearchInput, positive } : null,
            !isPrivateInstance ? { ...codySearchInput, positive } : null
        )
        setCodyFeedback(positive)
    }

    return (
        <aside
            role="region"
            aria-label="Search results information"
            className={classNames(props.className, styles.searchResultsInfoBar)}
            data-testid="results-info-bar"
        >
            {refFromCodySearch && codySearchInput.input && codySearchInput.translatedQuery === props.query ? (
                <Alert variant="info" className={styles.codyFeedbackAlert}>
                    Sourcegraph converted "<strong>{codySearchInput.input}</strong>" to "
                    <strong>{codySearchInput.translatedQuery}</strong>".{' '}
                    <small>
                        <Link target="blank" to="/help/code_search/reference/queries">
                            Complete query reference{' '}
                            <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                        </Link>
                    </small>
                    {codyFeedback === null ? (
                        <>
                            <Text className="my-2">Was this helpful?</Text>
                            <div>
                                <Button
                                    variant="secondary"
                                    outline={true}
                                    size="sm"
                                    onClick={() => collectCodyFeedback(true)}
                                >
                                    <Icon aria-hidden={true} className="mr-1" svgPath={mdiThumbUp} />
                                    Yes
                                </Button>
                                <Button
                                    className="ml-2"
                                    variant="secondary"
                                    outline={true}
                                    size="sm"
                                    onClick={() => collectCodyFeedback(false)}
                                >
                                    <Icon aria-hidden={true} className="mr-1" svgPath={mdiThumbDown} />
                                    No
                                </Button>
                            </div>
                        </>
                    ) : (
                        <Text className="my-2">
                            <strong>Thanks for your feedback!</strong>
                        </Text>
                    )}
                </Alert>
            ) : null}
            <div className={styles.row}>
                {props.stats}

                <div className={styles.expander} />

                <ul className="nav align-items-center">
                    <SearchActionsMenu
                        authenticatedUser={props.authenticatedUser}
                        createActions={createActions}
                        createCodeMonitorAction={createCodeMonitorAction}
                        canCreateMonitor={canCreateMonitorFromQuery}
                        results={props.results}
                        allExpanded={props.allExpanded}
                        onExpandAllResultsToggle={props.onExpandAllResultsToggle}
                        onSaveQueryClick={props.onSaveQueryClick}
                        onExportCsvClick={props.onExportCsvClick}
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
