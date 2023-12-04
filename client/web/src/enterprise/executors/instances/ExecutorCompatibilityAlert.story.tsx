import type { Decorator, StoryFn, Meta } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { ExecutorCompatibility } from '../../../graphql-operations'

import { ExecutorCompatibilityAlert } from './ExecutorCompatibilityAlert'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/executors/instances/ExecutorCompatibilityAlert',
    decorators: [decorator],
}

export default config

export const UpToDate: StoryFn = () => (
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

export const Outdated: StoryFn = () => (
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

export const VersionAhead: StoryFn = () => (
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
