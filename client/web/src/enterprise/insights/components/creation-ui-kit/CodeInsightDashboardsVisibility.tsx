import React from 'react'

import { Alert } from '@sourcegraph/wildcard'

export interface CodeInsightDashboardsVisibilityProps extends React.HTMLAttributes<HTMLDivElement> {
    dashboardCount: number
}

export const CodeInsightDashboardsVisibility: React.FunctionComponent<CodeInsightDashboardsVisibilityProps> = props => {
    const { dashboardCount, ...attributes } = props

    return (
        <Alert variant="note" {...attributes}>
            <h4 className="mt-0">This insight is included in {dashboardCount} other dashboards.</h4>
            <span className="text-muted">
                Changes to this insight will be shared across all instances of this insight.
            </span>
        </Alert>
    )
}
