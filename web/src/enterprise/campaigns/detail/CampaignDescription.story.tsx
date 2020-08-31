import { storiesOf } from '@storybook/react'
import React from 'react'
import { WebStory } from '../../../components/WebStory'
import webStyles from '../../../enterprise.scss'
import { CampaignDescription } from './CampaignDescription'

const { add } = storiesOf('web/campaigns/CampaignDescription', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('Overview', () => (
    <WebStory webStyles={webStyles}>
        {props => (
            <CampaignDescription
                {...props}
                description="This is an awesome campaign. It will do great things to your codebase."
            />
        )}
    </WebStory>
))
