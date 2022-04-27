import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../../../../../../components/WebStory'
import { InsightDashboard, InsightsDashboardOwnerType, InsightsDashboardType } from '../../../../../../../core/types'

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
        type: InsightsDashboardType.Custom,
        id: '101',
        title: 'Personal',
        insightIds: [],
        owners: [{ type: InsightsDashboardOwnerType.Personal, id: '101', title: 'Personal ' }],
    }

    return <EmptyInsightDashboard dashboard={dashboard} onAddInsight={noop} />
})
