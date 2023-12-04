import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { PerformanceGauge } from './PerformanceGauge'
import { StyledPerformanceGauge } from './story/StyledPerformanceGauge'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/site-admin/webhooks/PerformanceGauge',
    parameters: {
        chromatic: {
            viewports: [576],
        },
    },
    decorators: [decorator],
}

export default config

export const Loading: StoryFn = () => <WebStory>{() => <PerformanceGauge label="dog" />}</WebStory>

export const Zero: StoryFn = () => <WebStory>{() => <PerformanceGauge count={0} label="dog" />}</WebStory>

export const ZeroWithExplicitPlural: StoryFn = () => (
    <WebStory>{() => <PerformanceGauge count={0} label="wolf" plural="wolves" />}</WebStory>
)

ZeroWithExplicitPlural.storyName = 'zero with explicit plural'

export const One: StoryFn = () => <WebStory>{() => <PerformanceGauge count={1} label="dog" />}</WebStory>

export const Many: StoryFn = () => <WebStory>{() => <PerformanceGauge count={42} label="dog" />}</WebStory>

export const ManyWithExplicitPlural: StoryFn = () => (
    <WebStory>{() => <PerformanceGauge count={42} label="wolf" plural="wolves" />}</WebStory>
)

ManyWithExplicitPlural.storyName = 'many with explicit plural'

export const ClassOverrides: StoryFn = () => (
    <WebStory>{() => <StyledPerformanceGauge count={42} label="dog" />}</WebStory>
)

ClassOverrides.storyName = 'class overrides'
