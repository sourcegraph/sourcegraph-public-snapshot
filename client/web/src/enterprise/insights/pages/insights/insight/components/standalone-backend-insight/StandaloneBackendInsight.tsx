import React, { useCallback, useContext, useRef, useState } from 'react'

import classNames from 'classnames'
import { useHistory } from 'react-router'

import { asError } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Card, CardBody, useDebounce, useDeepMemo } from '@sourcegraph/wildcard'

import { InsightCard, InsightCardHeader, InsightCardLoading } from '../../../../../components'
import { FORM_ERROR, FormChangeEvent, SubmissionErrors } from '../../../../../components/form/hooks/useForm'
import {
    DrillDownInsightFilters,
    FilterSectionVisualMode,
    DrillDownInsightCreationForm,
    DrillDownFiltersStep,
    BackendInsightChart,
    BackendInsightErrorAlert,
    DrillDownFiltersFormValues,
    DrillDownInsightCreationFormValues,
} from '../../../../../components/insights-view-grid/components/backend-insight/components'
import { useInsightData } from '../../../../../components/insights-view-grid/hooks/use-insight-data'
import { ALL_INSIGHTS_DASHBOARD, BackendInsight, CodeInsightsBackendContext, InsightFilters } from '../../../../../core'
import { LazyQueryStatus } from '../../../../../hooks/use-parallel-requests/use-parallel-request'
import { getTrackingTypeByInsightType, useCodeInsightViewPings } from '../../../../../pings'
import { StandaloneInsightContextMenu } from '../context-menu/StandaloneInsightContextMenu'

import styles from './StandaloneBackendInsight.module.scss'

interface StandaloneBackendInsight extends TelemetryProps {
    insight: BackendInsight
    className?: string
}

export const StandaloneBackendInsight: React.FunctionComponent<StandaloneBackendInsight> = props => {
    const { telemetryService, insight, className } = props
    const history = useHistory()
    const { getBackendInsightData, createInsight, updateInsight } = useContext(CodeInsightsBackendContext)

    // Visual line chart settings
    const [zeroYAxisMin, setZeroYAxisMin] = useState(false)
    const [step, setStep] = useState(DrillDownFiltersStep.Filters)

    // Original insight filters values that are stored in setting subject with insight
    // configuration object, They are updated  whenever the user clicks update/save button
    const [originalInsightFilters, setOriginalInsightFilters] = useState(insight.filters)
    const insightCardReference = useRef<HTMLDivElement>(null)

    // Live valid filters from filter form. They are updated whenever the user is changing
    // filter value in filters fields.
    const [filters, setFilters] = useState<InsightFilters>(originalInsightFilters)
    const debouncedFilters = useDebounce(useDeepMemo<InsightFilters>(filters), 500)

    const { state, isVisible } = useInsightData(
        useCallback(() => getBackendInsightData({ ...insight, filters: debouncedFilters }), [
            insight,
            debouncedFilters,
            getBackendInsightData,
        ]),
        insightCardReference
    )

    const { trackMouseLeave, trackMouseEnter, trackDatumClicks } = useCodeInsightViewPings({
        telemetryService,
        insightType: getTrackingTypeByInsightType(insight.type),
    })

    const handleFilterChange = (event: FormChangeEvent<DrillDownFiltersFormValues>): void => {
        if (event.valid) {
            setFilters(event.values)
        }
    }

    const handleFilterSave = async (filters: InsightFilters): Promise<SubmissionErrors> => {
        try {
            await updateInsight({ insightId: insight.id, nextInsightData: { ...insight, filters } }).toPromise()
            setOriginalInsightFilters(filters)
            telemetryService.log('CodeInsightsSearchBasedFilterUpdating')
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    const handleInsightFilterCreation = async (
        values: DrillDownInsightCreationFormValues
    ): Promise<SubmissionErrors> => {
        try {
            await createInsight({
                insight: {
                    ...insight,
                    title: values.insightName,
                    filters,
                },
                dashboard: null,
            }).toPromise()

            setOriginalInsightFilters(filters)
            history.push(`/insights/dashboard${ALL_INSIGHTS_DASHBOARD.id}`)
            telemetryService.log('CodeInsightsSearchBasedFilterInsightCreation')
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    return (
        <div className={classNames(className, styles.root)}>
            <Card as={CardBody} className={styles.filters}>
                {step === DrillDownFiltersStep.Filters && (
                    <DrillDownInsightFilters
                        initialValues={filters}
                        originalValues={originalInsightFilters}
                        visualMode={FilterSectionVisualMode.HorizontalSections}
                        onFiltersChange={handleFilterChange}
                        onFilterSave={handleFilterSave}
                        onCreateInsightRequest={() => setStep(DrillDownFiltersStep.ViewCreation)}
                    />
                )}

                {step === DrillDownFiltersStep.ViewCreation && (
                    <DrillDownInsightCreationForm
                        onCreateInsight={handleInsightFilterCreation}
                        onCancel={() => setStep(DrillDownFiltersStep.Filters)}
                    />
                )}
            </Card>

            <InsightCard
                ref={insightCardReference}
                data-testid={`insight-standalone-card.${insight.id}`}
                className={styles.chart}
                onMouseEnter={trackMouseEnter}
                onMouseLeave={trackMouseLeave}
            >
                <InsightCardHeader title={insight.title}>
                    <StandaloneInsightContextMenu
                        insight={insight}
                        zeroYAxisMin={zeroYAxisMin}
                        onToggleZeroYAxisMin={setZeroYAxisMin}
                    />
                </InsightCardHeader>

                {state.status === LazyQueryStatus.Loading || !isVisible ? (
                    <InsightCardLoading>Loading code insight</InsightCardLoading>
                ) : state.status === LazyQueryStatus.Error ? (
                    <BackendInsightErrorAlert error={state.error} />
                ) : (
                    <BackendInsightChart {...state.data} locked={insight.isFrozen} onDatumClick={trackDatumClicks} />
                )}
            </InsightCard>
        </div>
    )
}
