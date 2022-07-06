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
import { SeriesDisplayOptionsInputRequired } from '../../../../core/types/insight/common'
import { getTrackingTypeByInsightType, useCodeInsightViewPings } from '../../../../pings'
import { FORM_ERROR, SubmissionErrors } from '../../../form/hooks/useForm'
import { InsightCard, InsightCardBanner, InsightCardHeader, InsightCardLoading } from '../../../views'
import { useVisibility } from '../../hooks/use-insight-data'
import { InsightContextMenu } from '../insight-context-menu/InsightContextMenu'
import { InsightContext } from '../InsightContext'

import {
    BackendInsightErrorAlert,
    DrillDownFiltersPopover,
    DrillDownInsightCreationFormValues,
    BackendInsightChart,
} from './components'
import { parseSeriesDisplayOptions } from './components/drill-down-filters-panel/drill-down-filters/utils'

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
    const { isVisible, wasEverVisible } = useVisibility(insightCardReference)

    // Use deep copy check in case if a setting subject has re-created copy of
    // the insight config with same structure and values. To avoid insight data
    // re-fetching.
    const cachedInsight = useDeepMemo(insight)

    // Original insight filters values that are stored in setting subject with insight
    // configuration object, They are updated  whenever the user clicks update/save button
    const [originalInsightFilters, setOriginalInsightFilters] = useState(cachedInsight.filters)
    const [originalSeriesDisplayOptions] = useState(cachedInsight.seriesDisplayOptions)

    // Live valid filters from filter form. They are updated whenever the user is changing
    // filter value in filters fields.
    const [filters, setFilters] = useState<InsightFilters>(originalInsightFilters)
    const [seriesDisplayOptions, setSeriesDisplayOptions] = useState(originalSeriesDisplayOptions)
    const [isFiltersOpen, setIsFiltersOpen] = useState(false)
    const debouncedFilters = useDebounce(useDeepMemo<InsightFilters>(filters), 500)

    const filterInput: InsightViewFiltersInput = {
        includeRepoRegex: debouncedFilters.includeRepoRegexp,
        excludeRepoRegex: debouncedFilters.excludeRepoRegexp,
        searchContexts: [debouncedFilters.context],
    }
    const displayInput: SeriesDisplayOptionsInput = {
        limit: seriesDisplayOptions?.limit,
        sortOptions: seriesDisplayOptions?.sortOptions,
    }

    const { error, loading, stopPolling } = useQuery<GetInsightViewResult, GetInsightViewVariables>(
        GET_INSIGHT_VIEW_GQL,
        {
            variables: { id: insight.id, filters: filterInput, seriesDisplayOptions: displayInput },
            fetchPolicy: 'cache-and-network',
            pollInterval: pollingInterval,
            skip: !wasEverVisible || (insightData && (!insightData.isFetchingHistoricalData || !isVisible)),
            context: { concurrentRequests: { key: 'GET_INSIGHT_VIEW' } },
            onCompleted: data => {
                const parsedData = createBackendInsightData({ ...insight, filters }, data.insightViews.nodes[0])
                if (!parsedData.isFetchingHistoricalData) {
                    stopPolling()
                }
                seriesToggleState.setSelectedSeriesIds([])
                setInsightData(parsedData)
            },
            onError: () => {
                stopPolling()
            },
        }
    )

    const handleFilterSave = async (
        filters: InsightFilters,
        displayOptions: SeriesDisplayOptionsInput
    ): Promise<SubmissionErrors> => {
        try {
            const insightWithNewFilters = { ...insight, filters, seriesDisplayOptions: displayOptions }

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
                seriesDisplayOptions,
            }

            await createInsight({
                insight: newInsight,
                dashboard: currentDashboard,
            }).toPromise()

            telemetryService.log('CodeInsightsSearchBasedFilterInsightCreation')
            setOriginalInsightFilters(filters)
            setSeriesDisplayOptions(originalSeriesDisplayOptions)
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

    const shareableUrl = `${window.location.origin}/insights/insight/${insight.id}`

    const handleSeriesDisplayOptionsChange = (options: SeriesDisplayOptionsInputRequired): void => {
        setSeriesDisplayOptions(options)
        seriesToggleState.setSelectedSeriesIds([])
    }

    return (
        <InsightCard
            {...otherProps}
            ref={mergedInsightCardReference}
            data-testid={`insight-card.${insight.id}`}
            className={classNames(otherProps.className, { [styles.cardWithFilters]: isFiltersOpen })}
            onMouseEnter={trackMouseEnter}
            onMouseLeave={trackMouseLeave}
        >
            <InsightCardHeader
                title={
                    <Link to={shareableUrl} target="_blank" rel="noopener noreferrer">
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
                            insight={insight}
                            onFilterChange={setFilters}
                            onFilterSave={handleFilterSave}
                            onInsightCreate={handleInsightFilterCreation}
                            onVisibilityChange={setIsFiltersOpen}
                            originalSeriesDisplayOptions={parseSeriesDisplayOptions(
                                insight.defaultSeriesDisplayOptions
                            )}
                            onSeriesDisplayOptionsChange={handleSeriesDisplayOptionsChange}
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
