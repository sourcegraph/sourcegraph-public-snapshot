import React from 'react'

import { TabList, type TabListProps, Tabs, type TabsProps } from '@sourcegraph/wildcard'

/** sourcegraph/wildcard `Tabs` with styling applied to prevent CLS on hovering the tabs. */
export const BatchChangeTabs: React.FunctionComponent<TabsProps> = props => (
    <Tabs size="medium" lazy={true} {...props} />
)

/** sourcegraph/wildcard `TabsList` with BC visual styling applied. */
export const BatchChangeTabList: React.FunctionComponent<TabListProps> = props => (
    <nav className="overflow-auto mb-2" aria-label="Batch Change">
        <TabList className="w-100 nav d-inline-flex d-sm-flex flex-nowrap text-nowrap" {...props} />
    </nav>
)
