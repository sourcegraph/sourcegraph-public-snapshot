import React from 'react'

import { mdiChartLineVariant } from '@mdi/js'

import { Badge, H1, Icon } from '@sourcegraph/wildcard'

export const AnalyticsPageTitle: React.FunctionComponent = ({ children }) => (
    <div className="d-flex flex-column justify-content-between align-items-start">
        <Badge variant="merged">Experimental</Badge>

        <H1 className="mb-4 mt-2 d-flex align-items-center">
            <Icon
                className="mr-1"
                color="var(--link-color)"
                svgPath={mdiChartLineVariant}
                size="sm"
                aria-label="Analytics icon"
            />
            {children}
        </H1>
    </div>
)
