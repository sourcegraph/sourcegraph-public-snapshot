import classNames from 'classnames'
import React, { Ref, useContext, useMemo, useRef, useState } from 'react'
import { useMergeRefs } from 'use-callback-ref'

import { ViewContexts, ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
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

function processData(data: ViewProviderResult | undefined): ViewProviderResult | undefined {

    // By possible types here data could be either undefined or error so
    // these checks should catch non-data and error like data cases
    if (!data || !data.view || isErrorLike(data.view)) {
        return data
    }

    // Here we iterate over all insight card content (usually we have just one
    // content within a card but by API it possible to have more than one chart within card)
    // So we Iterate over all content here
    const processedContent =  data.view.content.map(chartContent => {

        // This like says that if we got chart content (not a markup or custom content)
        // and this chart has the line type then we need to process this chart somehow
        if ('chart' in chartContent && chartContent.chart === 'line') {

            // We iterate over list of points here. These points looks like
            // { x: <timestamp>, [line1DataKey]: 100, [line2DataKey]: 200, ... }
            const processedData = chartContent.data.map(datum => {
                const processedDatum = {...datum}

                // We iterate over all series (lines that we have and by line dataKey
                // we access and change value of point object (datum)
                for (const line of chartContent.series) {
                    const { dataKey } = line

                    if (processedDatum[dataKey] !== null) {
                        processedDatum[dataKey] += 1000
                    }
                }

                return processedDatum
            })

            // Override original data (chartContent.data) with processedData object
            return { ...chartContent, data: processedData }
        }

        return chartContent
    })

    // Override original data.view with view with processed content object here
    return { ...data, view: { ...data.view, content: processedContent }  }
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

    // This is how we store any state in components like this
    // You can reed more about this here https://reactjs.org/docs/hooks-state.html
    const [isDataProcess, setProcessData] = useState<boolean>(false)

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

    // Here we're looking on our state and if users clicks on menu item and we set our state to
    // true then we should run processData function otherwise just return the original non-modified data
    const processedData = isDataProcess ? processData(data) : data

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
                        // Here we're passing state about data processing to context menu component
                        // to be able to check or uncheck context menu item
                        processedDataMode={isDataProcess}

                        // Here we're listening clicks on data process mode context menu item
                        onToggleDataProcessMode={() => setProcessData(state => !state)}
                    />
                )
            }
        >
            {resizing ? (
                <View.Banner>Resizing</View.Banner>
            ) : !processedData || loading || isDeleting || !isVisible ? (
                <View.LoadingContent text={isDeleting ? 'Deleting code insight' : 'Loading code insight'} />
            ) : isErrorLike(processedData.view) ? (
                <View.ErrorContent error={processedData.view} title={insight.id} />
            ) : (
                processedData.view && (
                    <LineChartSettingsContext.Provider value={{ zeroYAxisMin }}>
                        <View.Content
                            telemetryService={telemetryService}
                            content={processedData?.view.content}
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
