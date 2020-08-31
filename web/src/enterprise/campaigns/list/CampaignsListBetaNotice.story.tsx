import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { CampaignsListBetaNotice } from './CampaignsListBetaNotice'
import { WebStory } from '../../../components/WebStory'

const { add } = storiesOf('web/campaigns/CampaignsListBetaNotice', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('View', () => <WebStory webStyles={webStyles}>{props => <CampaignsListBetaNotice {...props} />}</WebStory>)
