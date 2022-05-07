import { number } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'

import { ExecutionStatsBar } from './ExecutionStatsBar'

const { add } = storiesOf('web/batches/batch-spec/execute', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('ExecutionStatsBar', () => (
    <WebStory>
        {props => (
            <ExecutionStatsBar
                {...props}
                errored={number('errored', 0)}
                ignored={number('ignored', 0)}
                queued={number('queued', 14)}
                processing={number('processing', 4)}
                completed={number('completed', 7)}
            />
        )}
    </WebStory>
))
