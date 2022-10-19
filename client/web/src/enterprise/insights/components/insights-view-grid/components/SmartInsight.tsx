import React, { forwardRef, useContext, useEffect, useState } from 'react'

import { useMergeRefs } from 'use-callback-ref'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useSearchParameters } from '@sourcegraph/wildcard'

import { CodeInsightsBackendContext, Insight, isBackendInsight } from '../../../core'

import { BackendInsightView } from './backend-insight'
import { BuiltInInsight } from './built-in-insight/BuiltInInsight'
import { InsightContextMenu } from './insight-context-menu/InsightContextMenu'
import { InsightContext } from './InsightContext'

export interface SmartInsightProps extends TelemetryProps, React.HTMLAttributes<HTMLElement> {
    insight: Insight
    resizing?: boolean
}

/**
 * Render smart insight with (gql or extension api) fetcher and independent mutation
 * actions.
 */
export const SmartInsight = forwardRef<HTMLElement, SmartInsightProps>((props, reference) => {
    const { insight, resizing = false, telemetryService, ...otherProps } = props

    const { currentDashboard, dashboards } = useContext(InsightContext)
    const { createInsight, updateInsight } = useContext(CodeInsightsBackendContext)

    const search = useSearchParameters()
    const mergedReference = useMergeRefs([reference])
    const [isZeroYAxisMin, setZeroYAxisMin] = useState(false)

    useEffect(() => {
        const insightIdToBeFocused = search.get('focused')
        const element = mergedReference.current

        if (element && insightIdToBeFocused === insight.id) {
            element.focus()
        }
    }, [insight.id, mergedReference, search])

    const handleCreateInsight = (insight: Insight): Promise<unknown> =>
        createInsight({ insight, dashboard: currentDashboard }).toPromise()

    const handleUpdateInsight = (insight: Insight): Promise<unknown> =>
        updateInsight({ insightId: insight.id, nextInsightData: insight }).toPromise()

    if (isBackendInsight(insight)) {
        return (
            <BackendInsightView
                ref={mergedReference}
                insight={insight}
                contextMenu={
                    <InsightContextMenu
                        insight={insight}
                        currentDashboard={currentDashboard}
                        dashboards={dashboards}
                        isZeroYAxisMin={isZeroYAxisMin}
                        onZeroYAxisMinChange={setZeroYAxisMin}
                    />
                }
                isZeroYAxisMin={isZeroYAxisMin}
                isResizing={resizing}
                telemetryService={telemetryService}
                onCreateInsight={handleCreateInsight}
                onUpdateInsight={handleUpdateInsight}
                {...otherProps}
            />
        )
    }

    // Lang-stats insight is handled by built-in fetchers
    return (
        <BuiltInInsight
            insight={insight}
            resizing={resizing}
            telemetryService={telemetryService}
            innerRef={mergedReference}
            {...otherProps}
        />
    )
})

SmartInsight.displayName = 'SmartInsight'
