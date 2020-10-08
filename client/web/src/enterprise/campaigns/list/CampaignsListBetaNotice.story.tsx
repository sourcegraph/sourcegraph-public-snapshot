import { storiesOf } from '@storybook/react'
import React from 'react'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { CampaignsListBetaNotice } from './CampaignsListBetaNotice'

const { add } = storiesOf('web/campaigns/CampaignsListBetaNotice', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('View', () => <EnterpriseWebStory>{props => <CampaignsListBetaNotice {...props} />}</EnterpriseWebStory>)
