import type { Meta, StoryFn, Decorator } from '@storybook/react'
import { subDays } from 'date-fns'

import { WebStory } from '../WebStory'

import { CreatedByAndUpdatedByInfoByline } from './CreatedByAndUpdatedByInfoByline'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/src/components/Byline',
    decorators: [decorator],
}

export default config

const THREE_DAYS_AGO = subDays(new Date(), 3).toISOString()

export const NeverUpdated: StoryFn = () => (
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

export const CreatedByDeletedUser: StoryFn = () => (
    <WebStory>
        {props => (
            <CreatedByAndUpdatedByInfoByline
                {...props}
                createdAt={THREE_DAYS_AGO}
                createdBy={null}
                updatedAt={THREE_DAYS_AGO}
                updatedBy={null}
            />
        )}
    </WebStory>
)

CreatedByDeletedUser.storyName = 'Created by deleted user'

export const NeverUpdatedSSBC: StoryFn = () => (
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

export const UpdatedSameUser: StoryFn = () => (
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

export const UpdatedDifferentUser: StoryFn = () => (
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

export const DatesWithoutAuthors: StoryFn = () => (
    <WebStory>
        {props => (
            <CreatedByAndUpdatedByInfoByline
                {...props}
                createdAt={THREE_DAYS_AGO}
                updatedAt={subDays(new Date(), 1).toISOString()}
                noAuthor={true}
            />
        )}
    </WebStory>
)

DatesWithoutAuthors.storyName = 'Created and updated dates without authors'
