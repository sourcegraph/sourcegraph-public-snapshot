import { useState } from 'react'

import { DecoratorFn, Story, Meta } from '@storybook/react'

import { WebStory } from '../../../../../../../components/WebStory'
import { InsightDashboard, InsightsDashboardOwnerType, InsightsDashboardType } from '../../../../../core'

import { DashboardSelect } from './DashboardSelect'

const decorator: DecoratorFn = story => <WebStory>{() => story()}</WebStory>

const config: Meta = {
    title: 'web/insights/DashboardSelect',
    decorators: [decorator],
    parameters: {
        chromatic: {
            viewports: [576, 1440],
        },
    },
}

export default config

const DASHBOARDS: InsightDashboard[] = [
    {
        type: InsightsDashboardType.Custom,
        id: '101',
        title: 'Personal',
        owners: [{ id: '101', title: 'Personal', type: InsightsDashboardOwnerType.Personal }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '102',
        title: 'Code Insights dashboard',
        owners: [{ id: '101', title: 'Personal', type: InsightsDashboardOwnerType.Personal }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '103',
        title: 'Experimental Insights dashboard',
        owners: [{ id: '101', title: 'Personal', type: InsightsDashboardOwnerType.Personal }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '104',
        title: 'Sourcegraph',
        owners: [{ id: '104', title: 'Sourcegraph', type: InsightsDashboardOwnerType.Personal }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '105',
        title: 'Loooong looo0000ong name of dashboard',
        owners: [{ id: '104', title: 'Sourcegraph', type: InsightsDashboardOwnerType.Personal }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '106',
        title: 'Loooong looo0000ong name of dashboard',
        owners: [{ id: '104', title: 'Sourcegraph', type: InsightsDashboardOwnerType.Personal }],
    },
]

export const DashboardSelectStory: Story = () => {
    const [value, setValue] = useState<string>()

    return <DashboardSelect value={value} dashboards={DASHBOARDS} onSelect={dashboard => setValue(dashboard.id)} />
}

DashboardSelectStory.storyName = 'DashboardSelect'
