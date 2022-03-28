import React, { useMemo } from 'react'

import { useLocation } from 'react-router'
import { of } from 'rxjs'

import { useObservable } from '@sourcegraph/wildcard'

import { ErrorBoundary } from '../../../../components/ErrorBoundary'
import { NotebookInsightsBlockProps } from '../../../../notebooks/blocks/insights/NotebooksInightsBlock'
import { SmartInsight } from '../../../insights/components/insights-view-grid/components/smart-insight/SmartInsight'
import { CodeInsightsBackendContext } from '../../../insights/core/backend/code-insights-backend-context'
import { useGetApi } from '../../../insights/hooks/use-get-api'

export const NotebookInsightsBlock: React.FunctionComponent<NotebookInsightsBlockProps> = React.memo(
    ({ id, telemetryService }) => {
        const location = useLocation()

        return (
            <ErrorBoundary location={location} render={error => <div>Error: {error.message}</div>}>
                <NotebookInsightsBlockInner id={id} telemetryService={telemetryService} />
            </ErrorBoundary>
        )
    }
)

const NotebookInsightsBlockInner: React.FunctionComponent<NotebookInsightsBlockProps> = ({
    id,
    telemetryService,
}: NotebookInsightsBlockProps) => {
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
