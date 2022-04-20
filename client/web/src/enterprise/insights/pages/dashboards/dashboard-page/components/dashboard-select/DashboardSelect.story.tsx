import { useState } from 'react'

import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../../../../../components/WebStory'
import { InsightDashboard, InsightsDashboardOwnerType, InsightsDashboardType } from '../../../../../core/types'

import { DashboardSelect } from './DashboardSelect'

const { add } = storiesOf('web/insights/DashboardSelect', module)
    .addDecorator(story => <WebStory>{() => story()}</WebStory>)
    .addParameters({
        chromatic: {
            viewports: [576, 1440],
        },
    })

const DASHBOARDS: InsightDashboard[] = [
    {
        type: InsightsDashboardType.Custom,
        id: '101',
        title: 'Personal',
        insightIds: [],
        owners: [{ id: '101', title: 'Personal', type: InsightsDashboardOwnerType.Personal }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '102',
        title: 'Code Insights dashboard',
        insightIds: [],
        owners: [{ id: '101', title: 'Personal', type: InsightsDashboardOwnerType.Personal }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '103',
        title: 'Experimental Insights dashboard',
        insightIds: [],
        owners: [{ id: '101', title: 'Personal', type: InsightsDashboardOwnerType.Personal }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '104',
        title: 'Sourcegraph',
        insightIds: [],
        owners: [{ id: '104', title: 'Sourcegraph', type: InsightsDashboardOwnerType.Personal }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '105',
        title: 'Loooong looo0000ong name of dashboard',
        insightIds: [],
        owners: [{ id: '104', title: 'Sourcegraph', type: InsightsDashboardOwnerType.Personal }],
    },
    {
        type: InsightsDashboardType.Custom,
        id: '106',
        title: 'Loooong looo0000ong name of dashboard',
        insightIds: [],
        owners: [{ id: '104', title: 'Sourcegraph', type: InsightsDashboardOwnerType.Personal }],
    },
]

add('DashboardSelect', () => {
    const [value, setValue] = useState<string>()

    return <DashboardSelect value={value} dashboards={DASHBOARDS} onSelect={dashboard => setValue(dashboard.id)} />
})
