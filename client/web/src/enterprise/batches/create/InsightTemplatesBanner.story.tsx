import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { InsightTemplatesBanner } from './InsightTemplatesBanner'

const { add } = storiesOf('web/batches/create', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            disableSnapshot: false,
        },
    })

add('InsightTemplatesBanner', () => (
    <WebStory>{props => <InsightTemplatesBanner {...props} insightTitle="My Go Insight" />}</WebStory>
))
