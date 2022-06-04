import React from 'react'

import { Alert, H4 } from '@sourcegraph/wildcard'

export interface CodeInsightDashboardsVisibilityProps extends React.HTMLAttributes<HTMLDivElement> {
    dashboardCount: number
}

export const CodeInsightDashboardsVisibility: React.FunctionComponent<
    React.PropsWithChildren<CodeInsightDashboardsVisibilityProps>
> = props => {
    const { dashboardCount, ...attributes } = props

    return (
        <Alert variant="note" {...attributes}>
            <H4 className="mt-0">This insight is included in {dashboardCount} other dashboards.</H4>
            <span className="text-muted">
                Changes to this insight will be shared across all instances of this insight.
            </span>
        </Alert>
    )
}
