import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { BatchSpecState } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../../components/WebStory'

import { BatchSpecStateBadge } from './BatchSpecStateBadge'

const decorator: Decorator = story => <div className="p-3">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute/BatchSpecStateBadge',
    decorators: [decorator],
}

export default config

export const BatchSpecStateBadgeStory: StoryFn = () => (
    <WebStory>
        {props => (
            <>
                {Object.values(BatchSpecState)
                    .sort()
                    .map(value => (
                        <div key={value} className="p-1">
                            <BatchSpecStateBadge state={value} {...props} />
                        </div>
                    ))}
            </>
        )}
    </WebStory>
)

BatchSpecStateBadgeStory.storyName = 'BatchSpecStateBadge'
