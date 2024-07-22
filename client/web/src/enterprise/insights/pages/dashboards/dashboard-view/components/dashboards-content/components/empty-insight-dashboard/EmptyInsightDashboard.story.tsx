import type { Decorator, StoryFn, Meta } from '@storybook/react'

import { WebStory } from '../../../../../../../../../components/WebStory'
import { type InsightDashboard, InsightsDashboardOwnerType, InsightsDashboardType } from '../../../../../../../core'

import { EmptyCustomDashboard } from './EmptyInsightDashboard'

const decorator: Decorator = story => <WebStory>{() => story()}</WebStory>

const config: Meta = {
    title: 'web/insights/EmptyInsightDashboard',
    decorators: [decorator],
    parameters: {},
}

export default config

export const EmptyInsightDashboardStory: StoryFn = () => {
    const dashboard: InsightDashboard = {
        type: InsightsDashboardType.Custom,
        id: '101',
        title: 'Personal',
        owners: [{ type: InsightsDashboardOwnerType.Personal, id: '101', title: 'Personal ' }],
    }

    return <EmptyCustomDashboard dashboard={dashboard} />
}

EmptyInsightDashboardStory.storyName = 'EmptyInsightDashboard'
