import { DecoratorFn, Meta, Story } from '@storybook/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../components/WebStory'

import { StatusMessagesNavItem } from './StatusMessagesNavItem'
import { allStatusMessages, newStatusMessageMock } from './StatusMessagesNavItem.mocks'

const decorator: DecoratorFn = Story => <Story />

const config: Meta = {
    title: 'web/nav/StatusMessagesNavItem',
    decorators: [decorator],
}

export default config

export const NoMessages: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={[newStatusMessageMock([])]}>
                <StatusMessagesNavItem disablePolling={true} />
            </MockedTestProvider>
        )}
    </WebStory>
)

NoMessages.storyName = 'No messages'

export const AllMessageTypes: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={[newStatusMessageMock(allStatusMessages)]}>
                <StatusMessagesNavItem disablePolling={true} />
            </MockedTestProvider>
        )}
    </WebStory>
)

AllMessageTypes.storyName = 'All message types'

export const IndexingMessage: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider
                mocks={[
                    newStatusMessageMock([
                        {
                            __typename: 'IndexingProgress',
                            indexed: 15,
                            notIndexed: 23,
                        },
                    ]),
                ]}
            >
                <StatusMessagesNavItem disablePolling={true} />
            </MockedTestProvider>
        )}
    </WebStory>
)

IndexingMessage.storyName = 'Indexing progress'

export const GitUpdatesDisabled: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider
                mocks={[
                    newStatusMessageMock([
                        {
                            __typename: 'GitUpdatesDisabled',
                            message: 'Repositories will not be cloned or updated.',
                        },
                    ]),
                ]}
            >
                <StatusMessagesNavItem disablePolling={true} />
            </MockedTestProvider>
        )}
    </WebStory>
)

GitUpdatesDisabled.storyName = 'Code syncing status'
