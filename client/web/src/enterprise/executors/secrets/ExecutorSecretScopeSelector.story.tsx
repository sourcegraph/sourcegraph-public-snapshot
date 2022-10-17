import { DecoratorFn, Meta, Story } from '@storybook/react'
import { WebStory } from '../../../components/WebStory'

import { ExecutorSecretScopeSelector } from './ExecutorSecretScopeSelector'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/executors/secrets/ExecutorSecretScopeSelector',
    parameters: {
        chromatic: {
            viewports: [576],
        },
    },
    decorators: [decorator],
}

export default config

export const Loading: Story = () => <WebStory>{() => <ExecutorSecretScopeSelector label="dog" />}</WebStory>

export const Zero: Story = () => <WebStory>{() => <ExecutorSecretScopeSelector count={0} label="dog" />}</WebStory>

export const ZeroWithExplicitPlural: Story = () => (
    <WebStory>{() => <ExecutorSecretScopeSelector count={0} label="wolf" plural="wolves" />}</WebStory>
)

ZeroWithExplicitPlural.storyName = 'zero with explicit plural'

export const One: Story = () => <WebStory>{() => <ExecutorSecretScopeSelector count={1} label="dog" />}</WebStory>

export const Many: Story = () => <WebStory>{() => <ExecutorSecretScopeSelector count={42} label="dog" />}</WebStory>

export const ManyWithExplicitPlural: Story = () => (
    <WebStory>{() => <ExecutorSecretScopeSelector count={42} label="wolf" plural="wolves" />}</WebStory>
)

ManyWithExplicitPlural.storyName = 'many with explicit plural'
