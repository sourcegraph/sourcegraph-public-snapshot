import { useState } from 'react'

import { DecoratorFn, Story, Meta } from '@storybook/react'

import { WebStory } from '../../../../../../../components/WebStory'
import { CustomInsightDashboard, InsightsDashboardOwnerType, InsightsDashboardType } from '../../../../../core'

import { DashboardSelect } from './DashboardSelect'

const decorator: DecoratorFn = story => <WebStory>{() => story()}</WebStory>

const config: Meta = {
    title: 'web/insights/DashboardSelect',
    decorators: [decorator],
}

export default config

const DASHBOARDS: CustomInsightDashboard[] = [
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
        owners: [{ id: '104', title: 'Sourcegraph', type: InsightsDashboardOwnerType.Organization }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '105',
        title: 'Loooong looo0000ong name of dashboard 1',
        owners: [{ id: '104', title: 'Sourcegraph', type: InsightsDashboardOwnerType.Organization }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '106',
        title: 'Loooong looo0000ong name of dashboard 2',
        owners: [{ id: '104', title: 'Sourcegraph', type: InsightsDashboardOwnerType.Organization }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '107',
        title: 'Global dashboard',
        owners: [{ id: '101', title: 'Personal', type: InsightsDashboardOwnerType.Global }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '108',
        title: 'Global FE dashboard',
        owners: [{ id: '101', title: 'Personal', type: InsightsDashboardOwnerType.Global }],
    },
]

export const DashboardSelectStory: Story = () => {
    const [dashboard, setDashboard] = useState<CustomInsightDashboard | undefined>()

    return (
        <section style={{ margin: '2rem' }}>
            <DashboardSelect dashboard={dashboard} dashboards={DASHBOARDS} onSelect={setDashboard} />
        </section>
    )
}

DashboardSelectStory.storyName = 'DashboardSelect'
