import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { Button } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../components/WebStory'

import { ReadOnlyBatchSpecAlert } from './ReadOnlyBatchSpecAlert'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute',
    decorators: [decorator],
}

export default config

export const ReadOnlyBatchSpecAlertStory: StoryFn = () => (
    <WebStory>
        {props => (
            <ReadOnlyBatchSpecAlert
                {...props}
                className="d-flex align-items-center pr-3"
                variant="info"
                header="This spec is read-only"
                message="We've preserved the original batch spec from this execution for you to inspect."
            >
                <Button variant="primary">Edit spec</Button>
            </ReadOnlyBatchSpecAlert>
        )}
    </WebStory>
)

ReadOnlyBatchSpecAlertStory.storyName = 'ReadOnlyBatchSpecAlert'
