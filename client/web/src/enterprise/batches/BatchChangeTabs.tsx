import React from 'react'

import { TabList, TabListProps, Tabs, TabsProps } from '@sourcegraph/wildcard'

import styles from './BatchChangeTabs.module.scss'

/** sourcegraph/wildcard `Tabs` with styling applied to prevent CLS on hovering the tabs. */
export const BatchChangeTabs: React.FunctionComponent<TabsProps> = props => (
    <Tabs className={styles.batchChangeTabs} lazy={true} {...props} />
)

/** sourcegraph/wildcard `TabsList` with BC visual styling applied. */
export const BatchChangeTabList: React.FunctionComponent<TabListProps> = props => (
    <div className="overflow-auto mb-2">
        <TabList className="w-100 nav d-inline-flex d-sm-flex flex-nowrap text-nowrap" {...props} />
    </div>
)
