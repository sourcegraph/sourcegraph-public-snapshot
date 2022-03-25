import React, { Ref, useCallback, useContext, useRef, useState } from 'react'

import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { asError, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useDebounce, Alert } from '@sourcegraph/wildcard'

import * as View from '../../../../../../views'
import { LineChartSettingsContext } from '../../../../../../views'
import { LockedChart } from '../../../../../../views/components/view/content/chart-view-content/charts/locked/LockedChart'
import { CodeInsightsBackendContext } from '../../../../core/backend/code-insights-backend-context'
import { InsightInProcessError } from '../../../../core/backend/utils/errors'
import { BackendInsight, InsightFilters } from '../../../../core/types'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight'
import { useDistinctValue } from '../../../../hooks/use-distinct-value'
import { useRemoveInsightFromDashboard } from '../../../../hooks/use-remove-insight'
import { DashboardInsightsContext } from '../../../../pages/dashboards/dashboard-page/components/dashboards-content/components/dashboard-inisghts/DashboardInsightsContext'
import { useCodeInsightViewPings, getTrackingTypeByInsightType } from '../../../../pings'
import { FORM_ERROR, SubmissionErrors } from '../../../form/hooks/useForm'
import { useInsightData } from '../../hooks/use-insight-data'
import { InsightContextMenu } from '../insight-context-menu/InsightContextMenu'

import { BackendAlertOverlay } from './BackendAlertOverlay'
import { DrillDownFiltersAction } from './components/drill-down-filters-action/DrillDownFiltersPanel'
import { DrillDownInsightCreationFormValues } from './components/drill-down-filters-panel/components/drill-down-insight-creation-form/DrillDownInsightCreationForm'
import { EMPTY_DRILLDOWN_FILTERS } from './components/drill-down-filters-panel/utils'

import styles from './BackendInsight.module.scss'

interface BackendInsightProps
    extends TelemetryProps,
        React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement> {
    insight: BackendInsight

    innerRef: Ref<HTMLElement>
    resizing?: boolean
}

/**
 * Renders BE search based insight. Fetches insight data by gql api handler.
 */
export const BackendInsightView: React.FunctionComponent<BackendInsightProps> = props => {
    const { telemetryService, insight, innerRef, resizing, ...otherProps } = props

    const { dashboard } = useContext(DashboardInsightsContext)
    const { getBackendInsightData, createInsight, updateInsight } = useContext(CodeInsightsBackendContext)

    // Visual line chart settings
    const [zeroYAxisMin, setZeroYAxisMin] = useState(false)
    const insightCardReference = useRef<HTMLDivElement>(null)
    const mergedInsightCardReference = useMergeRefs([insightCardReference, innerRef])

    // Use deep copy check in case if a setting subject has re-created copy of
    // the insight config with same structure and values. To avoid insight data
    // re-fetching.
    const cachedInsight = useDistinctValue(insight)

    // Original insight filters values that are stored in setting subject with insight
    // configuration object, They are updated  whenever the user clicks update/save button
    const [originalInsightFilters, setOriginalInsightFilters] = useState(
        cachedInsight.filters ?? EMPTY_DRILLDOWN_FILTERS
    )

    // Live valid filters from filter form. They are updated whenever the user is changing
    // filter value in filters fields.
    const [filters, setFilters] = useState<InsightFilters>(originalInsightFilters)
    const [isFiltersOpen, setIsFiltersOpen] = useState(false)
    const debouncedFilters = useDebounce(useDistinctValue<InsightFilters>(filters), 500)

    // Loading the insight backend data
    const { data, loading, error, isVisible } = useInsightData(
        useCallback(
            () =>
                getBackendInsightData({
                    ...cachedInsight,
                    filters: debouncedFilters,
                }),
            [cachedInsight, debouncedFilters, getBackendInsightData]
        ),
        insightCardReference
    )

    // Handle insight delete and remove actions
    const { loading: isDeleting, delete: handleDelete } = useDeleteInsight()
    const { remove: handleRemove, loading: isRemoving } = useRemoveInsightFromDashboard()

    const handleFilterSave = async (filters: InsightFilters): Promise<SubmissionErrors> => {
        try {
            const insightWithNewFilters = { ...insight, filters }

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

        if (!dashboard) {
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
                dashboard,
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
        <View.Root
            {...otherProps}
            title={insight.title}
            innerRef={mergedInsightCardReference}
            actions={
                isVisible && (
                    <>
                        <DrillDownFiltersAction
                            isOpen={isFiltersOpen}
                            popoverTargetRef={insightCardReference}
                            initialFiltersValue={filters}
                            originalFiltersValue={originalInsightFilters}
                            onFilterChange={setFilters}
                            onFilterSave={handleFilterSave}
                            onInsightCreate={handleInsightFilterCreation}
                            onVisibilityChange={setIsFiltersOpen}
                        />
                        <InsightContextMenu
                            insight={insight}
                            dashboard={dashboard}
                            menuButtonClassName="ml-1 d-inline-flex"
                            zeroYAxisMin={zeroYAxisMin}
                            onToggleZeroYAxisMin={() => setZeroYAxisMin(!zeroYAxisMin)}
                            onRemoveFromDashboard={dashboard => handleRemove({ insight, dashboard })}
                            onDelete={() => handleDelete(insight)}
                        />
                    </>
                )
            }
            data-testid={`insight-card.${insight.id}`}
            className={classNames(otherProps.className, { [styles.cardWithFilters]: isFiltersOpen })}
            onMouseEnter={trackMouseEnter}
            onMouseLeave={trackMouseLeave}
        >
            {resizing ? (
                <View.Banner>Resizing</View.Banner>
            ) : loading || isDeleting || !isVisible ? (
                <View.LoadingContent text={isDeleting ? 'Deleting code insight' : 'Loading code insight'} />
            ) : isRemoving ? (
                <View.LoadingContent text="Removing insight from the dashboard" />
            ) : isErrorLike(error) ? (
                <View.ErrorContent error={error} title={insight.id}>
                    {error instanceof InsightInProcessError ? (
                        <Alert className="m-0" variant="info">
                            {error.message}
                        </Alert>
                    ) : null}
                </View.ErrorContent>
            ) : insight.isFrozen ? (
                <LockedChart />
            ) : (
                data && (
                    <LineChartSettingsContext.Provider value={{ zeroYAxisMin }}>
                        <View.Content
                            content={data.view.content}
                            alert={
                                <BackendAlertOverlay
                                    hasNoData={!data.view.content.some(({ data }) => data.length > 0)}
                                    isFetchingHistoricalData={data.view.isFetchingHistoricalData}
                                />
                            }
                            onDatumLinkClick={trackDatumClicks}
                        />
                    </LineChartSettingsContext.Provider>
                )
            )}
            {
                // Passing children props explicitly to render any top-level content like
                // resize-handler from the react-grid-layout library
                isVisible && otherProps.children
            }
        </View.Root>
    )
}
