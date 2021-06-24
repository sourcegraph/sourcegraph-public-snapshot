import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import React from 'react'

import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { BatchSpecInfoByline } from './BatchSpecInfoByline'

const { add } = storiesOf('web/batches/preview/BatchSpecInfoByline', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('Default', () => (
    <EnterpriseWebStory>
        {() => (
            <BatchSpecInfoByline
                createdAt={subDays(new Date(), 3).toISOString()}
                creator={{ url: 'http://test.test/alice', username: 'alice' }}
            />
        )}
    </EnterpriseWebStory>
))
