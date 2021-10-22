import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { BatchChangeInfoByline } from './BatchChangeInfoByline'

const { add } = storiesOf('web/batches/BatchChangeInfoByline', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const THREE_DAYS_AGO = subDays(new Date(), 3).toISOString()

add('Never updated', () => (
    <WebStory>
        {props => (
            <BatchChangeInfoByline
                {...props}
                createdAt={THREE_DAYS_AGO}
                initialApplier={{ url: 'http://test.test/alice', username: 'alice' }}
                lastAppliedAt={THREE_DAYS_AGO}
                lastApplier={{ url: 'http://test.test/alice', username: 'alice' }}
            />
        )}
    </WebStory>
))

add('Updated (same user)', () => (
    <WebStory>
        {props => (
            <BatchChangeInfoByline
                {...props}
                createdAt={THREE_DAYS_AGO}
                initialApplier={{ url: 'http://test.test/alice', username: 'alice' }}
                lastAppliedAt={subDays(new Date(), 1).toISOString()}
                lastApplier={{ url: 'http://test.test/alice', username: 'alice' }}
            />
        )}
    </WebStory>
))

add('Updated (different users)', () => (
    <WebStory>
        {props => (
            <BatchChangeInfoByline
                {...props}
                createdAt={THREE_DAYS_AGO}
                initialApplier={{ url: 'http://test.test/alice', username: 'alice' }}
                lastAppliedAt={subDays(new Date(), 1).toISOString()}
                lastApplier={{ url: 'http://test.test/bob', username: 'bob' }}
            />
        )}
    </WebStory>
))
