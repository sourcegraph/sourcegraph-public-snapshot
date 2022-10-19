import { forwardRef, HTMLAttributes, ReactNode, useContext, useRef, useState } from 'react'

import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { asError } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link, useDebounce, useDeepMemo } from '@sourcegraph/wildcard'

import {
    SeriesDisplayOptionsInput,
    GetInsightDataResult,
    GetInsightDataVariables,
} from '../../../../../../graphql-operations'
import { useSeriesToggle } from '../../../../../../insights/utils/use-series-toggle'
import { BackendInsight, CodeInsightsBackendContext, InsightFilters } from '../../../../core'
import { getTrackingTypeByInsightType, useCodeInsightViewPings } from '../../../../pings'
import { FORM_ERROR, SubmissionErrors } from '../../../form'
import { InsightCard, InsightCardBanner, InsightCardHeader, InsightCardLoading } from '../../../views'
import { useVisibility } from '../../hooks/use-insight-data'
import { InsightContext } from '../InsightContext'

import {
    BackendInsightErrorAlert,
    DrillDownFiltersPopover,
    DrillDownInsightCreationFormValues,
    BackendInsightChart,
    parseSeriesLimit,
} from './components'
import { GET_INSIGHT_DATA } from './query'
import { parseBackendInsightResponse, insightPollingInterval } from './selectors'

import styles from './BackendInsight.module.scss'

interface BackendInsightProps extends TelemetryProps, Omit<HTMLAttributes<HTMLElement>, 'contextMenu'> {
    insight: BackendInsight
    contextMenu: ReactNode
    isZeroYAxisMin: boolean
    isResizing?: boolean
}

export const BackendInsightView = forwardRef<HTMLElement, BackendInsightProps>((props, ref) => {
    const { insight, contextMenu, isZeroYAxisMin, isResizing, children, telemetryService, ...otherProps } = props

    const { currentDashboard } = useContext(InsightContext)
    const { createInsight, updateInsight } = useContext(CodeInsightsBackendContext)

    const cardElementRef = useMergeRefs([ref])
    const { wasEverVisible, isVisible } = useVisibility(cardElementRef)

    const seriesToggleState = useSeriesToggle()

    // Original insight filters values that are stored in setting subject with insight
    // configuration object, They are updated  whenever the user clicks update/save button
    const [originalInsightFilters, setOriginalInsightFilters] = useState(insight.filters)

    // Live valid filters from filter form. They are updated whenever the user is changing
    // filter value in filters fields.
    const [filters, setFilters] = useState<InsightFilters>(originalInsightFilters)
    const [isFiltersOpen, setIsFiltersOpen] = useState(false)

    const debouncedFilters = useDebounce(useDeepMemo<InsightFilters>(filters), 500)

    const { data, error, loading, stopPolling, startPolling } = useQuery<GetInsightDataResult, GetInsightDataVariables>(
        GET_INSIGHT_DATA,
        {
            skip: !wasEverVisible,
            context: { concurrentRequests: { key: 'GET_INSIGHT_VIEW' } },
            variables: {
                id: insight.id,
                filters: {
                    includeRepoRegex: debouncedFilters.includeRepoRegexp,
                    excludeRepoRegex: debouncedFilters.excludeRepoRegexp,
                    searchContexts: [debouncedFilters.context],
                },
                seriesDisplayOptions: {
                    limit: parseSeriesLimit(debouncedFilters.seriesDisplayOptions.limit),
                    sortOptions: debouncedFilters.seriesDisplayOptions.sortOptions,
                },
            }
        }
    )

    const insightData = parseBackendInsightResponse({ ...insight, filters }, data)
    const isFetchingHistoricalData = insightData?.isInProgress
    const isPolling = useRef(false)

    // Not on the screen so stop polling if we are - multiple stop calls are safe
    if (error || !isVisible || !isFetchingHistoricalData) {
        isPolling.current = false
        stopPolling()
    } else if (isFetchingHistoricalData && !isPolling.current) {
        // we should start polling but multiple calls to startPolling reset the timer so
        // make sure we aren't already polling.
        isPolling.current = true
        startPolling(insightPollingInterval(insight))
    }

    const handleFiltersChange = (filters: InsightFilters): void => {
        seriesToggleState.setSelectedSeriesIds([])
        setFilters(filters)
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
            ref={cardElementRef}
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
                            anchor={cardElementRef}
                            initialFiltersValue={filters}
                            originalFiltersValue={originalInsightFilters}
                            onFilterChange={handleFiltersChange}
                            onFilterSave={handleFilterSave}
                            onInsightCreate={handleInsightFilterCreation}
                            onVisibilityChange={setIsFiltersOpen}
                        />
                        { contextMenu }
                    </>
                )}
            </InsightCardHeader>

            {isResizing ? (
                <InsightCardBanner>Resizing</InsightCardBanner>
            ) : error ? (
                <BackendInsightErrorAlert error={error} />
            ) : loading || !isVisible || !insightData ? (
                <InsightCardLoading>Loading code insight</InsightCardLoading>
            ) : (
                <BackendInsightChart
                    data={insightData.data}
                    seriesToggleState={seriesToggleState}
                    isInProgress={insightData.isInProgress}
                    isLocked={insight.isFrozen}
                    isZeroYAxisMin={isZeroYAxisMin}
                    onDatumClick={trackDatumClicks}
                />
            )}
            {
                // Passing children props explicitly to render any top-level content like
                // resize-handler from the react-grid-layout library
                isVisible && children
            }
        </InsightCard>
    )
})
