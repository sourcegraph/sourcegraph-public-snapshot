import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { InsightTemplatesBanner } from './InsightTemplatesBanner'

const { add } = storiesOf('web/batches/create/InsightTemplatesBanner', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('creating new batch change from insight', () => (
    <WebStory>{props => <InsightTemplatesBanner {...props} insightTitle="My Go Insight" type="create" />}</WebStory>
))

add('editing a batch spec from an insight template', () => (
    <WebStory>{props => <InsightTemplatesBanner {...props} insightTitle="My Go Insight" type="edit" />}</WebStory>
))
