import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../../../components/WebStory'
import { InsightDashboard, InsightsDashboardType } from '../../../../core/types'

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
        type: InsightsDashboardType.Personal,
        id: '101',
        title: 'Personal',
        builtIn: true,
        insightIds: [],
        owner: {
            id: '101',
            name: 'Pesonal',
        },
    },
    {
        type: InsightsDashboardType.Personal,
        id: '102',
        title: 'Code Insights dashboard',
        insightIds: [],
        owner: {
            id: '101',
            name: 'Pesonal',
        },
    },
    {
        type: InsightsDashboardType.Personal,
        id: '103',
        title: 'Experimental Insights dashboard',
        insightIds: [],
        owner: {
            id: '101',
            name: 'Pesonal',
        },
    },
    {
        type: InsightsDashboardType.Organization,
        id: '104',
        title: 'Sourcegraph',
        builtIn: true,
        insightIds: [],
        owner: {
            id: '104',
            name: 'Sourcegraph',
        },
    },
    {
        type: InsightsDashboardType.Organization,
        id: '105',
        title: 'Loooong looo0000ong name of dashboard',
        insightIds: [],
        owner: {
            id: '104',
            name: 'Sourcegraph',
        },
    },
    {
        type: InsightsDashboardType.Organization,
        id: '106',
        title: 'Loooong looo0000ong name of dashboard',
        insightIds: [],
        owner: {
            id: '104',
            name: 'Extended Sourcegraph space',
        },
    },
]

add('DashboardSelect', () => <DashboardSelect dashboards={DASHBOARDS} />)
