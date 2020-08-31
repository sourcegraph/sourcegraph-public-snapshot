import { storiesOf } from '@storybook/react'
import { boolean } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { CreateUpdateCampaignAlert } from './CreateUpdateCampaignAlert'
import { WebStory } from '../../../components/WebStory'

const { add } = storiesOf('web/campaigns/apply/CreateUpdateCampaignAlert', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('Create', () => (
    <WebStory webStyles={webStyles}>
        {props => (
            <CreateUpdateCampaignAlert
                {...props}
                specID="123"
                campaign={null}
                viewerCanAdminister={boolean('viewerCanAdminister', true)}
            />
        )}
    </WebStory>
))
add('Update', () => (
    <WebStory webStyles={webStyles}>
        {props => (
            <CreateUpdateCampaignAlert
                {...props}
                specID="123"
                campaign={{ id: '123', name: 'awesome-campaign', url: 'http://test.test/awesome' }}
                viewerCanAdminister={boolean('viewerCanAdminister', true)}
            />
        )}
    </WebStory>
))
