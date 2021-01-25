import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import React from 'react'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { SupersedingCampaignSpecAlert } from './SupersedingCampaignSpecAlert'

const { add } = storiesOf('web/campaigns/details/SupersedingCampaignSpecAlert', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('None published', () => (
    <EnterpriseWebStory>
        {() => (
            <SupersedingCampaignSpecAlert
                spec={{
                    applyURL: '/users/alice/campaigns/preview/123456SAMPLEID',
                    createdAt: subDays(new Date(), 1).toISOString(),
                }}
            />
        )}
    </EnterpriseWebStory>
))
