import { Meta, Story, DecoratorFn } from '@storybook/react'
import { subDays } from 'date-fns'

import { WebStory } from '../../../components/WebStory'

import { BatchChangeInfoByline } from './BatchChangeInfoByline'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/BatchChangeInfoByline',
    decorators: [decorator],
}

export default config

const THREE_DAYS_AGO = subDays(new Date(), 3).toISOString()

export const NeverUpdated: Story = () => (
    <WebStory>
        {props => (
            <BatchChangeInfoByline
                {...props}
                createdAt={THREE_DAYS_AGO}
                creator={{ url: 'http://test.test/alice', username: 'alice' }}
                lastAppliedAt={THREE_DAYS_AGO}
                lastApplier={{ url: 'http://test.test/alice', username: 'alice' }}
            />
        )}
    </WebStory>
)

NeverUpdated.storyName = 'Never updated'

export const NeverUpdatedSSBC: Story = () => (
    <WebStory>
        {props => (
            <BatchChangeInfoByline
                {...props}
                createdAt={THREE_DAYS_AGO}
                creator={{ url: 'http://test.test/alice', username: 'alice' }}
                lastAppliedAt={null}
                lastApplier={null}
            />
        )}
    </WebStory>
)

NeverUpdatedSSBC.storyName = 'Never updated (SSBC)'

export const UpdatedSameUser: Story = () => (
    <WebStory>
        {props => (
            <BatchChangeInfoByline
                {...props}
                createdAt={THREE_DAYS_AGO}
                creator={{ url: 'http://test.test/alice', username: 'alice' }}
                lastAppliedAt={subDays(new Date(), 1).toISOString()}
                lastApplier={{ url: 'http://test.test/alice', username: 'alice' }}
            />
        )}
    </WebStory>
)

UpdatedSameUser.storyName = 'Updated (same user)'

export const UpdatedDifferentUser: Story = () => (
    <WebStory>
        {props => (
            <BatchChangeInfoByline
                {...props}
                createdAt={THREE_DAYS_AGO}
                creator={{ url: 'http://test.test/alice', username: 'alice' }}
                lastAppliedAt={subDays(new Date(), 1).toISOString()}
                lastApplier={{ url: 'http://test.test/bob', username: 'bob' }}
            />
        )}
    </WebStory>
)

UpdatedDifferentUser.storyName = 'Updated (different users)'
