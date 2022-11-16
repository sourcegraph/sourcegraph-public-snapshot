import { Meta, Story, DecoratorFn } from '@storybook/react'
import { subDays } from 'date-fns'

import { WebStory } from '../WebStory'

import { CreatedByAndUpdatedByInfoByline } from './CreatedByAndUpdatedByInfoByline'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/src/components/Byline',
    decorators: [decorator],
}

export default config

const THREE_DAYS_AGO = subDays(new Date(), 3).toISOString()

export const NeverUpdated: Story = () => (
    <WebStory>
        {props => (
            <CreatedByAndUpdatedByInfoByline
                {...props}
                createdAt={THREE_DAYS_AGO}
                createdBy={{ url: 'http://test.test/alice', username: 'alice' }}
                updatedAt={THREE_DAYS_AGO}
                updatedBy={{ url: 'http://test.test/alice', username: 'alice' }}
            />
        )}
    </WebStory>
)

NeverUpdated.storyName = 'Never updated'

export const NeverUpdatedSSBC: Story = () => (
    <WebStory>
        {props => (
            <CreatedByAndUpdatedByInfoByline
                {...props}
                createdAt={THREE_DAYS_AGO}
                createdBy={{ url: 'http://test.test/alice', username: 'alice' }}
                updatedAt={null}
                updatedBy={null}
            />
        )}
    </WebStory>
)

NeverUpdatedSSBC.storyName = 'Never updated (SSBC)'

export const UpdatedSameUser: Story = () => (
    <WebStory>
        {props => (
            <CreatedByAndUpdatedByInfoByline
                {...props}
                createdAt={THREE_DAYS_AGO}
                createdBy={{ url: 'http://test.test/alice', username: 'alice' }}
                updatedAt={subDays(new Date(), 1).toISOString()}
                updatedBy={{ url: 'http://test.test/alice', username: 'alice' }}
            />
        )}
    </WebStory>
)

UpdatedSameUser.storyName = 'Updated (same user)'

export const UpdatedDifferentUser: Story = () => (
    <WebStory>
        {props => (
            <CreatedByAndUpdatedByInfoByline
                {...props}
                createdAt={THREE_DAYS_AGO}
                createdBy={{ url: 'http://test.test/alice', username: 'alice' }}
                updatedAt={subDays(new Date(), 1).toISOString()}
                updatedBy={{ url: 'http://test.test/bob', username: 'bob' }}
            />
        )}
    </WebStory>
)

UpdatedDifferentUser.storyName = 'Updated (different users)'
