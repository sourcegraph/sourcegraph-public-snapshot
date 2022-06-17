import { FunctionComponent, useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Insight, isBackendInsight } from '../../../../core'

import { StandaloneBackendInsight } from './standalone-backend-insight/StandaloneBackendInsight'
import { StandaloneRuntimeInsight } from './standalone-runtime-insight/StandaloneRuntimeInsight'

interface SmartStandaloneInsightProps extends TelemetryProps {
    insight: Insight
    className?: string
}

export const SmartStandaloneInsight: FunctionComponent<SmartStandaloneInsightProps> = props => {
    const { insight, telemetryService, className } = props

    useEffect(() => {
        telemetryService.logPageView('Insights')
    }, [telemetryService, insight.id])

    if (isBackendInsight(insight)) {
        return <StandaloneBackendInsight insight={insight} telemetryService={telemetryService} className={className} />
    }

    // Search based extension and lang stats insight are handled by built-in fetchers
    return <StandaloneRuntimeInsight insight={insight} telemetryService={telemetryService} className={className} />
}
