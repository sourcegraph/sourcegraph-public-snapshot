import React from 'react'

import { InstanceData } from '../model'
import { InstanceListItem } from './InstanceListItem'
import { Tab, TabList, Tabs } from '@sourcegraph/wildcard'
import { ContactSupport } from '../common/ContactSupport'

export const InstanceList: React.FunctionComponent<{
    instances: InstanceData[]
    className?: string
}> = ({ instances, className }) => (
    <div className={className}>
        <Tabs>
            <TabList wrapperClassName="border-bottom-0">
                <Tab>
                    <span className="tablist-wrapper--tab-label">Created by you</span>
                </Tab>
            </TabList>
        </Tabs>
        {instances.length === 0 ? (
            <p>No instances found.</p>
        ) : (
            <ol className="list-group">
                {instances.map((instance, i) => (
                    <InstanceListItem key={i} instance={instance} tag="li" className="list-group-item p-3" />
                ))}
            </ol>
        )}
        <ContactSupport className="mt-3" />
    </div>
)
