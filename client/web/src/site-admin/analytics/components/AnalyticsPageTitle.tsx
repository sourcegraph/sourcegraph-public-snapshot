import React from 'react'

import { mdiChartLineVariant, mdiInformationOutline } from '@mdi/js'

import { Badge, H1, Icon, Tooltip } from '@sourcegraph/wildcard'

export const AnalyticsPageTitle: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => (
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
            <Tooltip content="Data is updated every 24 hour.">
                <Icon
                    className="ml-1"
                    svgPath={mdiInformationOutline}
                    aria-label="Analytics info icon"
                    size="sm"
                    color="var(--link-color)"
                />
            </Tooltip>
        </H1>
    </div>
)
