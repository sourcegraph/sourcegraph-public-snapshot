import { storiesOf } from '@storybook/react'
import React from 'react'
import { CampaignInfoByline } from './CampaignInfoByline'
import { subDays } from 'date-fns'
import { EnterpriseWebStory } from '../../../components/WebStory'

const { add } = storiesOf('web/campaigns/CampaignInfoByline', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('Default', () => (
    <EnterpriseWebStory>
        {props => (
            <CampaignInfoByline
                {...props}
                createdAt={subDays(new Date(), 3).toISOString()}
                initialApplier={{ url: 'http://test.test/alice', username: 'alice' }}
                lastAppliedAt={subDays(new Date(), 1).toISOString()}
                lastApplier={{ url: 'http://test.test/bob', username: 'bob' }}
            />
        )}
    </EnterpriseWebStory>
))
