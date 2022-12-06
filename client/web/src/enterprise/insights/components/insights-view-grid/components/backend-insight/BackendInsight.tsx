import { forwardRef, HTMLAttributes, useContext, useRef, useState } from 'react'

import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { isDefined } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link, useDebounce, useDeepMemo } from '@sourcegraph/wildcard'

import {
    SeriesDisplayOptionsInput,
    GetInsightViewResult,
    GetInsightViewVariables,
} from '../../../../../../graphql-operations'
import { useSeriesToggle } from '../../../../../../insights/utils/use-series-toggle'
import { BackendInsight, BackendInsightData, CodeInsightsBackendContext, InsightFilters } from '../../../../core'
import { GET_INSIGHT_VIEW_GQL } from '../../../../core/backend/gql-backend'
import { createBackendInsightData } from '../../../../core/backend/gql-backend/methods/get-backend-insight-data/deserializators'
import { insightPollingInterval } from '../../../../core/backend/gql-backend/utils/insight-polling'
import { useSaveInsightAsNewView } from '../../../../core/hooks/use-save-insight-as-new-view'
import { getTrackingTypeByInsightType, useCodeInsightViewPings } from '../../../../pings'
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
import { BackendInsightTimeoutIcon } from './components/backend-insight-chart/BackendInsightChart'

import styles from './BackendInsight.module.scss'

interface BackendInsightProps extends TelemetryProps, HTMLAttributes<HTMLElement> {
    insight: BackendInsight
    resizing?: boolean
}

export const BackendInsightView = forwardRef<HTMLElement, BackendInsightProps>((props, ref) => {
    const { telemetryService, insight, resizing, children, className, ...attributes } = props

    const { currentDashboard, dashboards } = useContext(InsightContext)
    const { updateInsight } = useContext(CodeInsightsBackendContext)
    const [saveNewView] = useSaveInsightAsNewView({ dashboard: currentDashboard })

    const cardElementRef = useMergeRefs([ref])
    const { wasEverVisible, isVisible } = useVisibility(cardElementRef)

    const seriesToggleState = useSeriesToggle()

    // Original insight filters values that are stored in setting subject with insight
    // configuration object, They are updated  whenever the user clicks update/save button
    const [originalInsightFilters, setOriginalInsightFilters] = useState(insight.filters)

    // Live valid filters from filter form. They are updated whenever the user is changing
    // filter value in filters fields.
    const [filters, setFilters] = useState<InsightFilters>(originalInsightFilters)
    const [insightData, setInsightData] = useState<BackendInsightData | undefined>()
    const [zeroYAxisMin, setZeroYAxisMin] = useState(false)
    const [isFiltersOpen, setIsFiltersOpen] = useState(false)

    const debouncedFilters = useDebounce(useDeepMemo<InsightFilters>(filters), 500)

    const { error, loading, stopPolling, startPolling } = useQuery<GetInsightViewResult, GetInsightViewVariables>(
        GET_INSIGHT_VIEW_GQL,
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
            },
            onCompleted: data => {
                // This query requests a list of 1 insight view if there is an error and the insightView
                // will be null and error is populated
                const node = data.insightViews.nodes[0]

                seriesToggleState.setSelectedSeriesIds([])
                setInsightData(isDefined(node) ? createBackendInsightData({ ...insight, filters }, node) : undefined)
            },
        }
    )

    const isFetchingHistoricalData = insightData?.isFetchingHistoricalData
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

    async function handleFilterSave(filters: InsightFilters): Promise<void> {
        const seriesDisplayOptions: SeriesDisplayOptionsInput = {
            limit: parseSeriesLimit(filters.seriesDisplayOptions.limit),
            sortOptions: filters.seriesDisplayOptions.sortOptions,
        }
        const insightWithNewFilters = { ...insight, filters, seriesDisplayOptions }

        await updateInsight({ insightId: insight.id, nextInsightData: insightWithNewFilters }).toPromise()

        telemetryService.log('CodeInsightsSearchBasedFilterUpdating')
        setOriginalInsightFilters(filters)
        setIsFiltersOpen(false)
    }

    const handleInsightFilterCreation = async (values: DrillDownInsightCreationFormValues): Promise<void> => {
        const { insightName } = values

        if (!currentDashboard) {
            return
        }

        await saveNewView({
            insight,
            filters,
            title: insightName,
            dashboard: currentDashboard,
        })

        telemetryService.log('CodeInsightsSearchBasedFilterInsightCreation')
        setOriginalInsightFilters(filters)
        setIsFiltersOpen(false)
    }

    const { trackMouseLeave, trackMouseEnter, trackDatumClicks } = useCodeInsightViewPings({
        telemetryService,
        insightType: getTrackingTypeByInsightType(insight.type),
    })

    return (
        <InsightCard
            {...attributes}
            ref={cardElementRef}
            data-testid={`insight-card.${insight.id}`}
            aria-label={`${insight.title} insight`}
            role="listitem"
            className={classNames(className, { [styles.cardWithFilters]: isFiltersOpen })}
            onMouseEnter={trackMouseEnter}
            onMouseLeave={trackMouseLeave}
        >
            <InsightCardHeader
                title={
                    <Link
                        to={`${window.location.origin}/insights/insight/${insight.id}`}
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        {insight.title}
                    </Link>
                }
            >
                {isVisible && (
                    <>
                        {insightData?.isAllSeriesErrored && <BackendInsightTimeoutIcon timeoutLevel="insight" />}
                        <DrillDownFiltersPopover
                            isOpen={isFiltersOpen}
                            anchor={cardElementRef}
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
                isVisible && children
            }
        </InsightCard>
    )
})
