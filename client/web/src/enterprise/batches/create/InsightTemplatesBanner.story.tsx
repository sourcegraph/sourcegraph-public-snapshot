import type { Meta, Story, DecoratorFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { InsightTemplatesBanner } from './InsightTemplatesBanner'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/create/InsightTemplatesBanner',
    decorators: [decorator],
}

export default config

export const CreatingNewBatchChangeFromInsight: Story = () => (
    <WebStory>{props => <InsightTemplatesBanner {...props} insightTitle="My Go Insight" type="create" />}</WebStory>
)

CreatingNewBatchChangeFromInsight.storyName = 'Creating new batch change from insight'

export const EditingBatchSpecFromInsightTemplate: Story = () => (
    <WebStory>{props => <InsightTemplatesBanner {...props} insightTitle="My Go Insight" type="edit" />}</WebStory>
)

EditingBatchSpecFromInsightTemplate.storyName = 'Editing a batch spec from an insight template'
