import React from 'react'

import { Alert } from '@sourcegraph/wildcard'

export interface CodeInsightDashboardsVisibilityProps extends React.HTMLAttributes<HTMLDivElement> {
    dashboardCount: number
}

export const CodeInsightDashboardsVisibility: React.FunctionComponent<CodeInsightDashboardsVisibilityProps> = props => {
    const { dashboardCount, ...attributes } = props

    return (
        <Alert variant="secondary" {...attributes}>
            <span className="font-weight-bold">This insight is included in {dashboardCount} other dashboards.</span>
            <br />
            <span>Changes to this insight will be shared across all instances of this insight.</span>
        </Alert>
    )
}
