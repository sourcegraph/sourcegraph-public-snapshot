import React from 'react'

import { mdiChartLineVariant } from '@mdi/js'

import { SiteAdminPageTitle } from '../../components/SiteAdminPageTitle'

export const AnalyticsPageTitle: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => (
    <SiteAdminPageTitle icon={{ svgPath: mdiChartLineVariant, ariaLabel: 'Analytics icon' }}>
        <span>Analytics</span>
        {children}
    </SiteAdminPageTitle>
)
