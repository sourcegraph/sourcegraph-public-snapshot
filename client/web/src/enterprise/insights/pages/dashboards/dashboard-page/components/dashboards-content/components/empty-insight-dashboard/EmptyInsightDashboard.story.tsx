import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { WebStory } from '../../../../../../../../../components/WebStory'
import { InsightDashboard, InsightsDashboardType } from '../../../../../../../core/types'

import { EmptyInsightDashboard } from './EmptyInsightDashboard'

const { add } = storiesOf('web/insights/EmptyInsightDashboard', module)
    .addDecorator(story => <WebStory>{() => story()}</WebStory>)
    .addParameters({
        chromatic: {
            viewports: [576, 1440],
        },
    })

add('EmptyInsightDashboard', () => {
    const dashboard: InsightDashboard = {
        type: InsightsDashboardType.Personal,
        id: '101',
        title: 'Personal',
        builtIn: true,
        insightIds: [],
        owner: {
            id: '101',
            name: 'Pesonal',
        },
        settingsKey: 'test',
    }

    return <EmptyInsightDashboard dashboard={dashboard} onAddInsight={noop} />
})
