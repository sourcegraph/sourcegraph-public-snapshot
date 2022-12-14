import { FC, useContext, useMemo } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { useExperimentalFeatures } from '../../../../../../../../../stores'
import { SmartInsightsViewGrid, InsightContext } from '../../../../../../../components'
import { CodeInsightsBackendContext, InsightDashboard } from '../../../../../../../core'
import { EmptyInsightDashboard } from '../empty-insight-dashboard/EmptyInsightDashboard'

interface DashboardInsightsProps extends TelemetryProps {
    currentDashboard: InsightDashboard
    className?: string
    onAddInsightRequest?: () => void
}

export const DashboardInsights: FC<DashboardInsightsProps> = props => {
    const { telemetryService, currentDashboard, className, onAddInsightRequest } = props

    const { getInsights } = useContext(CodeInsightsBackendContext)
    const { codeInsightsCompute = false } = useExperimentalFeatures()

    const insights = useObservable(
        useMemo(
            () =>
                getInsights({
                    dashboardId: currentDashboard.id,
                    withCompute: codeInsightsCompute,
                }),
            [getInsights, codeInsightsCompute, currentDashboard.id]
        )
    )

    const insightContextValue = useMemo(() => ({ currentDashboard }), [currentDashboard])

    if (insights === undefined) {
        return <LoadingSpinner aria-hidden={true} inline={false} />
    }

    return (
        <InsightContext.Provider value={insightContextValue}>
            {insights.length > 0 ? (
                <SmartInsightsViewGrid insights={insights} telemetryService={telemetryService} className={className} />
            ) : (
                <EmptyInsightDashboard dashboard={currentDashboard} onAddInsightRequest={onAddInsightRequest} />
            )}
        </InsightContext.Provider>
    )
}
