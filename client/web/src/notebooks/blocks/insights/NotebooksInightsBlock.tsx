import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

export interface NotebookInsightsBlockProps extends TelemetryProps {
    id: string
}

export const NotebookInsightsBlock: React.FunctionComponent<NotebookInsightsBlockProps> = React.memo(() => (
    <div>Code Insights not enabled</div>
))
