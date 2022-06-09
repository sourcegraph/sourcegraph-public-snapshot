import { storiesOf } from '@storybook/react'

import { Button } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../components/WebStory'

import { ReadOnlyBatchSpecAlert } from './ReadOnlyBatchSpecAlert'

storiesOf('web/batches/batch-spec/execute', module)
    .addDecorator(story => <div className="container p-3">{story()}</div>)
    .add('ReadOnlyBatchSpecAlert', () => (
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
    ))
