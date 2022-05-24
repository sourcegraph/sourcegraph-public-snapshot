import React, { Ref, useCallback, useContext, useRef, useState } from 'react'

import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { asError } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useDebounce, useDeepMemo } from '@sourcegraph/wildcard'

import { BackendInsight, CodeInsightsBackendContext, InsightFilters } from '../../../../core'
import { LazyQueryStatus } from '../../../../hooks/use-parallel-requests/use-parallel-request'
import { getTrackingTypeByInsightType, useCodeInsightViewPings } from '../../../../pings'
import { FORM_ERROR, SubmissionErrors } from '../../../form/hooks/useForm'
import { InsightCard, InsightCardBanner, InsightCardHeader, InsightCardLoading } from '../../../views'
import { useInsightData } from '../../hooks/use-insight-data'
import { InsightContextMenu } from '../insight-context-menu/InsightContextMenu'
import { InsightContext } from '../InsightContext'

import {
    BackendInsightErrorAlert,
    DrillDownFiltersPopover,
    DrillDownInsightCreationFormValues,
    BackendInsightChart,
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
 * Renders BE search based insight. Fetches insight data by gql api handler.
 */
export const BackendInsightView: React.FunctionComponent<React.PropsWithChildren<BackendInsightProps>> = props => {
    const { telemetryService, insight, innerRef, resizing, ...otherProps } = props

    const { currentDashboard, dashboards } = useContext(InsightContext)
    const { getBackendInsightData, createInsight, updateInsight } = useContext(CodeInsightsBackendContext)

    // Visual line chart settings
    const [zeroYAxisMin, setZeroYAxisMin] = useState(false)
    const insightCardReference = useRef<HTMLDivElement>(null)
    const mergedInsightCardReference = useMergeRefs([insightCardReference, innerRef])

    // Use deep copy check in case if a setting subject has re-created copy of
    // the insight config with same structure and values. To avoid insight data
    // re-fetching.
    const cachedInsight = useDeepMemo(insight)

    // Original insight filters values that are stored in setting subject with insight
    // configuration object, They are updated  whenever the user clicks update/save button
    const [originalInsightFilters, setOriginalInsightFilters] = useState(cachedInsight.filters)

    // Live valid filters from filter form. They are updated whenever the user is changing
    // filter value in filters fields.
    const [filters, setFilters] = useState<InsightFilters>(originalInsightFilters)
    const [isFiltersOpen, setIsFiltersOpen] = useState(false)
    const debouncedFilters = useDebounce(useDeepMemo<InsightFilters>(filters), 500)

    // Loading the insight backend data
    const { state, isVisible } = useInsightData(
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
            className={classNames(otherProps.className, { [styles.cardWithFilters]: isFiltersOpen })}
            onMouseEnter={trackMouseEnter}
            onMouseLeave={trackMouseLeave}
        >
            <InsightCardHeader title={insight.title}>
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
            ) : state.status === LazyQueryStatus.Loading || !isVisible ? (
                <InsightCardLoading>Loading code insight</InsightCardLoading>
            ) : state.status === LazyQueryStatus.Error ? (
                <BackendInsightErrorAlert error={state.error} />
            ) : (
                <BackendInsightChart {...state.data} locked={insight.isFrozen} onDatumClick={trackDatumClicks} />
            )}
            {
                // Passing children props explicitly to render any top-level content like
                // resize-handler from the react-grid-layout library
                isVisible && otherProps.children
            }
        </InsightCard>
    )
}
