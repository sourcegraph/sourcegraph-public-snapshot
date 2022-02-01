import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { WebStory } from '../../../../../../../../../components/WebStory'
import { InsightDashboard, InsightsDashboardScope, InsightsDashboardType } from '../../../../../../../core/types'

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
        scope: InsightsDashboardScope.Personal,
        type: InsightsDashboardType.BuiltIn,
        id: '101',
        title: 'Personal',
        insightIds: [],
        owner: {
            id: '101',
            name: 'Pesonal',
        },
    }

    return <EmptyInsightDashboard dashboard={dashboard} onAddInsight={noop} />
})
