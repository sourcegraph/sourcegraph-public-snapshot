import { storiesOf } from '@storybook/react'
import React from 'react'
import { WebStory } from '../../../components/WebStory'
import webStyles from '../../../enterprise.scss'
import { CreateCampaignPage } from './CreateCampaignPage'

const { add } = storiesOf('web/campaigns/CreateCampaignPage', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('Page', () => <WebStory webStyles={webStyles}>{props => <CreateCampaignPage {...props} />}</WebStory>)
