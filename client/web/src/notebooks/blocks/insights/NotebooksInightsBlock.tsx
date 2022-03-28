import React, { useMemo } from 'react'

import { of } from 'rxjs'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/wildcard'

import { SmartInsight } from '../../../enterprise/insights/components/insights-view-grid/components/smart-insight/SmartInsight'
import { CodeInsightsBackendContext } from '../../../enterprise/insights/core/backend/code-insights-backend-context'
import { useGetApi } from '../../../enterprise/insights/hooks/use-get-api'

interface NotebookInsightsBlockProps extends TelemetryProps {
    id: string
}

export const NotebookInsightsBlock: React.FunctionComponent<NotebookInsightsBlockProps> = React.memo(
    ({ id, telemetryService }) => {
        const api = useGetApi()

        const insight = useObservable(useMemo(() => (api ? api.getInsightById(id) : of(null)), [api, id]))

        return insight && api ? (
            <CodeInsightsBackendContext.Provider value={api}>
                <SmartInsight insight={insight} telemetryService={telemetryService} style={{ height: '300px' }} />
            </CodeInsightsBackendContext.Provider>
        ) : (
            <div>No insight</div>
        )
    }
)
