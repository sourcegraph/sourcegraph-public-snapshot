import React, { Ref, useContext, useRef, useState } from 'react'

import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { asError } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link, useDebounce, useDeepMemo } from '@sourcegraph/wildcard'

import { useFeatureFlag } from '../../../../../../featureFlags/useFeatureFlag'
import {
    InsightViewFiltersInput,
    SeriesDisplayOptionsInput,
    GetInsightViewResult,
    GetInsightViewVariables,
} from '../../../../../../graphql-operations'
import { useSeriesToggle } from '../../../../../../insights/utils/use-series-toggle'
import { BackendInsight, BackendInsightData, CodeInsightsBackendContext, InsightFilters } from '../../../../core'
import { GET_INSIGHT_VIEW_GQL } from '../../../../core/backend/gql-backend'
import { createBackendInsightData } from '../../../../core/backend/gql-backend/methods/get-backend-insight-data/deserializators'
import { insightPollingInterval } from '../../../../core/backend/gql-backend/utils/insight-polling'
import { getTrackingTypeByInsightType, useCodeInsightViewPings } from '../../../../pings'
import { FORM_ERROR, SubmissionErrors } from '../../../form'
import { InsightCard, InsightCardBanner, InsightCardHeader, InsightCardLoading } from '../../../views'
import { useVisibility } from '../../hooks/use-insight-data'
import { InsightContextMenu } from '../insight-context-menu/InsightContextMenu'
import { InsightContext } from '../InsightContext'

import {
    BackendInsightErrorAlert,
    DrillDownFiltersPopover,
    DrillDownInsightCreationFormValues,
    BackendInsightChart,
    parseSeriesLimit,
} from './components'

import styles from './BackendInsight.module.scss'

interface BackendInsightProps
    extends TelemetryProps,
        React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement> {
    insight: BackendInsight

    innerRef: Ref<HTMLElement>
    resizing?: boolean
}

/**
 * Renders search based insight. Fetches insight data by gql api handler.
 */
export const BackendInsightView: React.FunctionComponent<React.PropsWithChildren<BackendInsightProps>> = props => {
    const { telemetryService, insight, innerRef, resizing, ...otherProps } = props

    const { currentDashboard, dashboards } = useContext(InsightContext)
    const { createInsight, updateInsight } = useContext(CodeInsightsBackendContext)
    // seriesToggleState is instantiated at this level to prevent the state from being
    // deleted when the insight is scrolled out of view
    const seriesToggleState = useSeriesToggle()
    const [insightData, setInsightData] = useState<BackendInsightData | undefined>()
    const [enablePolling] = useFeatureFlag('insight-polling-enabled', true)
    const pollingInterval = enablePolling ? insightPollingInterval(insight) : 0

    // Visual line chart settings
    const [zeroYAxisMin, setZeroYAxisMin] = useState(false)
    const insightCardReference = useRef<HTMLDivElement>(null)
    const mergedInsightCardReference = useMergeRefs([insightCardReference, innerRef])
    const { wasEverVisible, isVisible } = useVisibility(insightCardReference)

    // Original insight filters values that are stored in setting subject with insight
    // configuration object, They are updated  whenever the user clicks update/save button
    const [originalInsightFilters, setOriginalInsightFilters] = useState(insight.filters)

    // Live valid filters from filter form. They are updated whenever the user is changing
    // filter value in filters fields.
    const [filters, setFilters] = useState<InsightFilters>(originalInsightFilters)
    const [isFiltersOpen, setIsFiltersOpen] = useState(false)
    const debouncedFilters = useDebounce(useDeepMemo<InsightFilters>(filters), 500)

    const filterInput: InsightViewFiltersInput = {
        includeRepoRegex: debouncedFilters.includeRepoRegexp,
        excludeRepoRegex: debouncedFilters.excludeRepoRegexp,
        searchContexts: [debouncedFilters.context],
    }
    const seriesDisplayOptions: SeriesDisplayOptionsInput = {
        limit: parseSeriesLimit(debouncedFilters.seriesDisplayOptions.limit),
        sortOptions: debouncedFilters.seriesDisplayOptions.sortOptions,
    }

    const { error, loading, stopPolling, startPolling } = useQuery<GetInsightViewResult, GetInsightViewVariables>(
        GET_INSIGHT_VIEW_GQL,
        {
            variables: { id: insight.id, filters: filterInput, seriesDisplayOptions },
            fetchPolicy: 'cache-and-network',
            skip: !wasEverVisible,
            context: { concurrentRequests: { key: 'GET_INSIGHT_VIEW' } },
            onCompleted: data => {
                const parsedData = createBackendInsightData({ ...insight, filters }, data.insightViews.nodes[0])
                seriesToggleState.setSelectedSeriesIds([])
                setInsightData(parsedData)
            },
        }
    )

    const isFetchingHistoricalData = insightData?.isFetchingHistoricalData
    const isPolling = useRef(false)

    // polling is disabled ignore all
    if (enablePolling) {
        // not on the screen so stop polling if we are - multiple stop calls are safe
        if (error || !isVisible || !isFetchingHistoricalData) {
            isPolling.current = false
            stopPolling()
        } else if (isFetchingHistoricalData && !isPolling.current) {
            // we should start polling but multiple calls to startPolling reset the timer so
            // make sure we aren't already polling.
            isPolling.current = true
            startPolling(pollingInterval)
        }
    }

    async function handleFilterSave(filters: InsightFilters): Promise<SubmissionErrors> {
        try {
            const seriesDisplayOptions: SeriesDisplayOptionsInput = {
                limit: parseSeriesLimit(filters.seriesDisplayOptions.limit),
                sortOptions: filters.seriesDisplayOptions.sortOptions,
            }
            const insightWithNewFilters = { ...insight, filters, seriesDisplayOptions }

            await updateInsight({ insightId: insight.id, nextInsightData: insightWithNewFilters }).toPromise()

            telemetryService.log('CodeInsightsSearchBasedFilterUpdating')

            setOriginalInsightFilters(filters)
            setIsFiltersOpen(false)
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    const handleInsightFilterCreation = async (
        values: DrillDownInsightCreationFormValues
    ): Promise<SubmissionErrors> => {
        const { insightName } = values

        if (!currentDashboard) {
            return
        }

        try {
            const newInsight = {
                ...insight,
                title: insightName,
                filters,
            }

            await createInsight({
                insight: newInsight,
                dashboard: currentDashboard,
            }).toPromise()

            telemetryService.log('CodeInsightsSearchBasedFilterInsightCreation')
            setOriginalInsightFilters(filters)
            setIsFiltersOpen(false)
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    const { trackMouseLeave, trackMouseEnter, trackDatumClicks } = useCodeInsightViewPings({
        telemetryService,
        insightType: getTrackingTypeByInsightType(insight.type),
    })

    return (
        <InsightCard
            {...otherProps}
            ref={mergedInsightCardReference}
            data-testid={`insight-card.${insight.id}`}
            aria-label="Insight card"
            className={classNames(otherProps.className, { [styles.cardWithFilters]: isFiltersOpen })}
            onMouseEnter={trackMouseEnter}
            onMouseLeave={trackMouseLeave}
        >
            <InsightCardHeader
                title={
                    <Link
                        to={`${window.location.origin}/insights/insight/${insight.id}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        aria-label="Go to the insight page"
                    >
                        {insight.title}
                    </Link>
                }
            >
                {isVisible && (
                    <>
                        <DrillDownFiltersPopover
                            isOpen={isFiltersOpen}
                            anchor={insightCardReference}
                            initialFiltersValue={filters}
                            originalFiltersValue={originalInsightFilters}
                            onFilterChange={setFilters}
                            onFilterSave={handleFilterSave}
                            onInsightCreate={handleInsightFilterCreation}
                            onVisibilityChange={setIsFiltersOpen}
                        />
                        <InsightContextMenu
                            insight={insight}
                            currentDashboard={currentDashboard}
                            dashboards={dashboards}
                            zeroYAxisMin={zeroYAxisMin}
                            onToggleZeroYAxisMin={() => setZeroYAxisMin(!zeroYAxisMin)}
                        />
                    </>
                )}
            </InsightCardHeader>

            {resizing ? (
                <InsightCardBanner>Resizing</InsightCardBanner>
            ) : error ? (
                <BackendInsightErrorAlert error={error} />
            ) : loading || !isVisible || !insightData ? (
                <InsightCardLoading>Loading code insight</InsightCardLoading>
            ) : (
                <BackendInsightChart
                    {...insightData}
                    locked={insight.isFrozen}
                    zeroYAxisMin={zeroYAxisMin}
                    seriesToggleState={seriesToggleState}
                    onDatumClick={trackDatumClicks}
                />
            )}
            {
                // Passing children props explicitly to render any top-level content like
                // resize-handler from the react-grid-layout library
                isVisible && otherProps.children
            }
        </InsightCard>
    )
}
