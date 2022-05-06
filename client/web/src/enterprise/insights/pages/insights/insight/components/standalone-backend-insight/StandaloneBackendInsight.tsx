import React, { useCallback, useContext, useRef, useState } from 'react'

import classNames from 'classnames'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useDebounce, useDeepMemo } from '@sourcegraph/wildcard'

import { InsightCard, InsightCardHeader, InsightCardLoading } from '../../../../../components'
import {
    BackendInsightChart,
    BackendInsightErrorAlert,
} from '../../../../../components/insights-view-grid/components/backend-insight/components'
import { useInsightData } from '../../../../../components/insights-view-grid/hooks/use-insight-data'
import { BackendInsight, CodeInsightsBackendContext, InsightFilters } from '../../../../../core'
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
    const { getBackendInsightData } = useContext(CodeInsightsBackendContext)

    // Visual line chart settings
    const [zeroYAxisMin, setZeroYAxisMin] = useState(false)

    // Original insight filters values that are stored in setting subject with insight
    // configuration object, They are updated  whenever the user clicks update/save button
    const [originalInsightFilters] = useState(insight.filters)
    const insightCardReference = useRef<HTMLDivElement>(null)

    // Live valid filters from filter form. They are updated whenever the user is changing
    // filter value in filters fields.
    const [filters] = useState<InsightFilters>(originalInsightFilters)
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

    return (
        <div className={classNames(className, styles.root)}>
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
