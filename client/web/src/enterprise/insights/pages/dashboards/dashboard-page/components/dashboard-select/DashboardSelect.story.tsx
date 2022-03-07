import { storiesOf } from '@storybook/react'
import React, { useState } from 'react'

import { AuthenticatedUser } from '../../../../../../../auth'
import { WebStory } from '../../../../../../../components/WebStory'
import { InsightDashboard, InsightsDashboardScope, InsightsDashboardType } from '../../../../../core/types'

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
        scope: InsightsDashboardScope.Personal,
        type: InsightsDashboardType.BuiltIn,
        id: '101',
        title: 'Personal',
        insightIds: [],
        owner: {
            id: '101',
            name: 'Pesonal',
        },
    },
    {
        scope: InsightsDashboardScope.Personal,
        type: InsightsDashboardType.Custom,
        id: '102',
        title: 'Code Insights dashboard',
        insightIds: [],
        owner: {
            id: '101',
            name: 'Pesonal',
        },
    },
    {
        scope: InsightsDashboardScope.Personal,
        type: InsightsDashboardType.Custom,
        id: '103',
        title: 'Experimental Insights dashboard',
        insightIds: [],
        owner: {
            id: '101',
            name: 'Pesonal',
        },
    },
    {
        scope: InsightsDashboardScope.Organization,
        type: InsightsDashboardType.BuiltIn,
        id: '104',
        title: 'Sourcegraph',
        insightIds: [],
        owner: {
            id: '104',
            name: 'Sourcegraph',
        },
    },
    {
        scope: InsightsDashboardScope.Organization,
        type: InsightsDashboardType.Custom,
        id: '105',
        title: 'Loooong looo0000ong name of dashboard',
        insightIds: [],
        owner: {
            id: '104',
            name: 'Sourcegraph',
        },
    },
    {
        scope: InsightsDashboardScope.Organization,
        type: InsightsDashboardType.Custom,
        id: '106',
        title: 'Loooong looo0000ong name of dashboard',
        insightIds: [],
        owner: {
            id: '104',
            name: 'Extended Sourcegraph space',
        },
    },
]

const USER: Partial<AuthenticatedUser> = {
    organizations: {
        nodes: [
            {
                id: '1',
                name: 'Sourcegraph',
                displayName: 'Sourcegraph',
                url: 'https://sourcegraph.com',
                settingsURL: 'https://sourcegraph.com/settings',
            },
        ],
    },
}

add('DashboardSelect', () => {
    const [value, setValue] = useState<string>()

    return (
        <DashboardSelect
            value={value}
            dashboards={DASHBOARDS}
            user={USER as AuthenticatedUser}
            onSelect={dashboard => setValue(dashboard.id)}
        />
    )
})
