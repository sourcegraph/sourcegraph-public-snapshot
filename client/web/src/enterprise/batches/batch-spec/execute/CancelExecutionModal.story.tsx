import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../components/WebStory'

import { CancelExecutionModal } from './CancelExecutionModal'

const { add } = storiesOf('web/batches/batch-spec/execute', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('CancelExecutionModal', () => (
    <WebStory>
        {props => (
            <CancelExecutionModal
                {...props}
                modalBody={<p>Are you sure you want to cancel the current execution?</p>}
                isOpen={true}
                isLoading={boolean('isLoading', false)}
                onCancel={noop}
                onConfirm={noop}
            />
        )}
    </WebStory>
))
