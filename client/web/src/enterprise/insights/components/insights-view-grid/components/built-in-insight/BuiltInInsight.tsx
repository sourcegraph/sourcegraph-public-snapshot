import classNames from 'classnames'
import React, { Ref, useContext, useMemo, useRef, useState } from 'react'
import { useMergeRefs } from 'use-callback-ref'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import * as View from '../../../../../../views'
import { LineChartSettingsContext } from '../../../../../../views'
import { CodeInsightsBackendContext } from '../../../../core/backend/code-insights-backend-context'
import { LangStatsInsight } from '../../../../core/types'
import { SearchExtensionBasedInsight } from '../../../../core/types/insight/search-insight'
import { useDeleteInsight } from '../../../../hooks/use-delete-insight'
import { useDistinctValue } from '../../../../hooks/use-distinct-value'
import { DashboardInsightsContext } from '../../../../pages/dashboards/dashboard-page/components/dashboards-content/components/dashboard-inisghts/DashboardInsightsContext'
import { useInsightData } from '../../hooks/use-insight-data'
import { InsightContextMenu } from '../insight-context-menu/InsightContextMenu'

interface BuiltInInsightProps<D extends keyof ViewContexts> extends TelemetryProps, React.HTMLAttributes<HTMLElement> {
    insight: SearchExtensionBasedInsight | LangStatsInsight
    where: D
    context: ViewContexts[D]
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
export function BuiltInInsight<D extends keyof ViewContexts>(props: BuiltInInsightProps<D>): React.ReactElement {
    const { insight, resizing, telemetryService, where, context, innerRef, ...otherProps } = props
    const { getBuiltInInsightData } = useContext(CodeInsightsBackendContext)
    const { dashboard } = useContext(DashboardInsightsContext)

    const insightCardReference = useRef<HTMLDivElement>(null)
    const mergedInsightCardReference = useMergeRefs([insightCardReference, innerRef])

    const cachedInsight = useDistinctValue(insight)

    const { data, loading, isVisible } = useInsightData(
        useMemo(() => () => getBuiltInInsightData({ insight: cachedInsight, options: { where, context } }), [
            getBuiltInInsightData,
            cachedInsight,
            where,
            context,
        ]),
        insightCardReference
    )

    // Visual line chart settings
    const [zeroYAxisMin, setZeroYAxisMin] = useState(false)
    const { delete: handleDelete, loading: isDeleting } = useDeleteInsight()

    return (
        <View.Root
            {...otherProps}
            innerRef={mergedInsightCardReference}
            data-testid={`insight-card.${insight.id}`}
            title={insight.title}
            className={classNames('extension-insight-card', otherProps.className)}
            actions={
                isVisible && (
                    <InsightContextMenu
                        insight={insight}
                        dashboard={dashboard}
                        menuButtonClassName="ml-1 d-inline-flex"
                        zeroYAxisMin={zeroYAxisMin}
                        onToggleZeroYAxisMin={() => setZeroYAxisMin(!zeroYAxisMin)}
                        onDelete={() => handleDelete(insight)}
                    />
                )
            }
        >
            {resizing ? (
                <View.Banner>Resizing</View.Banner>
            ) : !data || loading || isDeleting || !isVisible ? (
                <View.LoadingContent text={isDeleting ? 'Deleting code insight' : 'Loading code insight'} />
            ) : isErrorLike(data.view) ? (
                <View.ErrorContent error={data.view} title={insight.id} />
            ) : (
                data.view && (
                    <LineChartSettingsContext.Provider value={{ zeroYAxisMin }}>
                        <View.Content
                            telemetryService={telemetryService}
                            content={data.view.content}
                            viewTrackingType={insight.viewType}
                            containerClassName="extension-insight-card"
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
