import { storiesOf } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { PerformanceGauge } from './PerformanceGauge'
import { StyledPerformanceGauge } from './story/StyledPerformanceGauge'

const { add } = storiesOf('web/site-admin/webhooks/PerformanceGauge', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [576],
        },
    })

add('loading', () => <WebStory>{() => <PerformanceGauge label="dog" />}</WebStory>)
add('zero', () => <WebStory>{() => <PerformanceGauge count={0} label="dog" />}</WebStory>)
add('zero with explicit plural', () => (
    <WebStory>{() => <PerformanceGauge count={0} label="wolf" plural="wolves" />}</WebStory>
))
add('one', () => <WebStory>{() => <PerformanceGauge count={1} label="dog" />}</WebStory>)
add('many', () => <WebStory>{() => <PerformanceGauge count={42} label="dog" />}</WebStory>)
add('many with explicit plural', () => (
    <WebStory>{() => <PerformanceGauge count={42} label="wolf" plural="wolves" />}</WebStory>
))
add('class overrides', () => <WebStory>{() => <StyledPerformanceGauge count={42} label="dog" />}</WebStory>)
