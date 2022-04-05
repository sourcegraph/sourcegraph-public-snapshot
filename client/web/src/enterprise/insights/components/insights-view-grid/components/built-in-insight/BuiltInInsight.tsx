import React, { Ref, useContext, useMemo, useRef, useState } from 'react'

import { useMergeRefs } from 'use-callback-ref'

import { isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import * as View from '../../../../../../views'
import { LineChartSettingsContext } from '../../../../../../views'
import { CodeInsightsBackendContext } from '../../../../core/backend/code-insights-backend-context'
import { LangStatsInsight } from '../../../../core/types'
import { SearchRuntimeBasedInsight } from '../../../../core/types/insight/types/search-insight'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight'
import { useDistinctValue } from '../../../../hooks/use-distinct-value'
import { useRemoveInsightFromDashboard } from '../../../../hooks/use-remove-insight'
import { DashboardInsightsContext } from '../../../../pages/dashboards/dashboard-page/components/dashboards-content/components/dashboard-inisghts/DashboardInsightsContext'
import { useCodeInsightViewPings, getTrackingTypeByInsightType } from '../../../../pings'
import { useInsightData } from '../../hooks/use-insight-data'
import { InsightContextMenu } from '../insight-context-menu/InsightContextMenu'

interface BuiltInInsightProps extends TelemetryProps, React.HTMLAttributes<HTMLElement> {
    insight: SearchRuntimeBasedInsight | LangStatsInsight
    innerRef: Ref<HTMLElement>
    resizing: boolean
}

/**
 * Historically we had a few insights that were worked via extension API
 * search-based, code-stats insight
 *
 * This component renders insight card that works almost like before with extensions
 * Component sends FE network request to get and process information but does that in
 * main work thread instead of using Extension API.
 */
export function BuiltInInsight(props: BuiltInInsightProps): React.ReactElement {
    const { insight, resizing, telemetryService, innerRef, ...otherProps } = props
    const { getBuiltInInsightData } = useContext(CodeInsightsBackendContext)
    const { dashboard } = useContext(DashboardInsightsContext)

    const insightCardReference = useRef<HTMLDivElement>(null)
    const mergedInsightCardReference = useMergeRefs([insightCardReference, innerRef])

    const cachedInsight = useDistinctValue(insight)

    const { data, loading, isVisible } = useInsightData(
        useMemo(() => () => getBuiltInInsightData({ insight: cachedInsight }), [getBuiltInInsightData, cachedInsight]),
        insightCardReference
    )

    // Visual line chart settings
    const [zeroYAxisMin, setZeroYAxisMin] = useState(false)
    const { delete: handleDelete, loading: isDeleting } = useDeleteInsight()
    const { remove: handleRemove, loading: isRemoving } = useRemoveInsightFromDashboard()

    const { trackDatumClicks, trackMouseLeave, trackMouseEnter } = useCodeInsightViewPings({
        telemetryService,
        insightType: getTrackingTypeByInsightType(insight.type),
    })

    return (
        <View.Root
            {...otherProps}
            innerRef={mergedInsightCardReference}
            title={insight.title}
            actions={
                isVisible && (
                    <InsightContextMenu
                        insight={insight}
                        dashboard={dashboard}
                        menuButtonClassName="ml-1 d-inline-flex"
                        zeroYAxisMin={zeroYAxisMin}
                        onToggleZeroYAxisMin={() => setZeroYAxisMin(!zeroYAxisMin)}
                        onRemoveFromDashboard={dashboard => handleRemove({ insight, dashboard })}
                        onDelete={() => handleDelete(insight)}
                    />
                )
            }
            data-testid={`insight-card.${insight.id}`}
            onMouseEnter={trackMouseEnter}
            onMouseLeave={trackMouseLeave}
        >
            {resizing ? (
                <View.Banner>Resizing</View.Banner>
            ) : !data || loading || isDeleting || !isVisible ? (
                <View.LoadingContent text={isDeleting ? 'Deleting code insight' : 'Loading code insight'} />
            ) : isRemoving ? (
                <View.LoadingContent text="Removing insight from the dashboard" />
            ) : isErrorLike(data.view) ? (
                <View.ErrorContent error={data.view} title={insight.id} />
            ) : (
                data.view && (
                    <LineChartSettingsContext.Provider value={{ zeroYAxisMin }}>
                        <View.Content
                            content={data.view.content}
                            onDatumLinkClick={trackDatumClicks}
                            locked={insight.isFrozen}
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
