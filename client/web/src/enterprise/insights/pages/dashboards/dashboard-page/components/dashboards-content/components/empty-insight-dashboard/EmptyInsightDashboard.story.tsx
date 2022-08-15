import { DecoratorFn, Story, Meta } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../../../../../../components/WebStory'
import { InsightDashboard, InsightsDashboardOwnerType, InsightsDashboardType } from '../../../../../../../core/types'

import { EmptyInsightDashboard } from './EmptyInsightDashboard'

const decorator: DecoratorFn = story => <WebStory>{() => story()}</WebStory>

const config: Meta = {
    title: 'web/insights/EmptyInsightDashboard',
    decorators: [decorator],
    parameters: {
        chromatic: {
            viewports: [576, 1440],
        },
    },
}

export default config

export const EmptyInsightDashboardStory: Story = () => {
    const dashboard: InsightDashboard = {
        type: InsightsDashboardType.Custom,
        id: '101',
        title: 'Personal',
        insightIds: [],
        owners: [{ type: InsightsDashboardOwnerType.Personal, id: '101', title: 'Personal ' }],
    }

    return <EmptyInsightDashboard dashboard={dashboard} onAddInsight={noop} />
}

EmptyInsightDashboardStory.storyName = 'EmptyInsightDashboard'
