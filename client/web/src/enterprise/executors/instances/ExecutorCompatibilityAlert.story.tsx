import { DecoratorFn, Story, Meta } from '@storybook/react'

import { ExecutorCompatibility } from '@sourcegraph/search'

import { WebStory } from '../../../components/WebStory'

import { ExecutorCompatibilityAlert } from './ExecutorCompatibilityAlert'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/executors/instances/ExecutorCompatibilityAlert',
    decorators: [decorator],
}

export default config

export const UpToDate: Story = () => (
    <WebStory>
        {props => (
            <ExecutorCompatibilityAlert
                {...props}
                compatibility={ExecutorCompatibility.UP_TO_DATE}
                hostname="executor1.sgdev.org"
            />
        )}
    </WebStory>
)

// This story is expected to be empty.
UpToDate.storyName = 'Executor is up to date'

export const Outdated: Story = () => (
    <WebStory>
        {props => (
            <ExecutorCompatibilityAlert
                {...props}
                compatibility={ExecutorCompatibility.OUTDATED}
                hostname="executor1.sgdev.org"
            />
        )}
    </WebStory>
)

Outdated.storyName = 'Executor is outdated'

export const VersionAhead: Story = () => (
    <WebStory>
        {props => (
            <ExecutorCompatibilityAlert
                {...props}
                compatibility={ExecutorCompatibility.VERSION_AHEAD}
                hostname="executor1.sgdev.org"
            />
        )}
    </WebStory>
)

VersionAhead.storyName = 'Executor is a version ahead'
