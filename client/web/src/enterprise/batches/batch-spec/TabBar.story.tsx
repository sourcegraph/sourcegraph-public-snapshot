import { select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { BatchSpecState } from '../../../graphql-operations'

import { TabBar } from './TabBar'

const { add } = storiesOf('web/batches/batch-spec/TabBar', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('creating a new batch change', () => (
    <WebStory>
        {props => <TabBar {...props} batchChange={null} batchSpec={null} activeTabName="configuration" />}
    </WebStory>
))

const USER_NAMESPACE = {
    __typename: 'User',
    username: 'my-username',
    displayName: 'My User',
    viewerCanAdminister: false,
    url: 'some/url/to/user',
    id: 'test-1234',
} as const

add('editing unexecuted batch spec', () => (
    <WebStory>
        {props => (
            <TabBar
                {...props}
                batchChange={{ name: 'my-batch-change', namespace: USER_NAMESPACE }}
                batchSpec={{ id: '1234', state: BatchSpecState.PENDING, applyURL: null, startedAt: null }}
                activeTabName={select('Active tab', ['configuration', 'batch spec'], 'batch spec')}
            />
        )}
    </WebStory>
))

add('executing a batch spec', () => (
    <WebStory>
        {props => (
            <TabBar
                {...props}
                batchChange={{ name: 'my-batch-change', namespace: USER_NAMESPACE }}
                batchSpec={{
                    id: '1234',
                    state: BatchSpecState.PROCESSING,
                    applyURL: null,
                    startedAt: new Date().toISOString(),
                }}
                activeTabName={select('Active tab', ['configuration', 'batch spec', 'execution'], 'execution')}
            />
        )}
    </WebStory>
))

add('previewing changesets', () => (
    <WebStory>
        {props => (
            <TabBar
                {...props}
                batchChange={{ name: 'my-batch-change', namespace: USER_NAMESPACE }}
                batchSpec={{
                    id: '1234',
                    state: BatchSpecState.COMPLETED,
                    applyURL: '/some/url/to/apply',
                    startedAt: new Date().toISOString(),
                }}
                activeTabName={select('Active tab', ['configuration', 'batch spec', 'execution', 'preview'], 'preview')}
            />
        )}
    </WebStory>
))
