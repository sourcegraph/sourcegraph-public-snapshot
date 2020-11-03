import { storiesOf } from '@storybook/react'
import React from 'react'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { CreateCampaignPage } from './CreateCampaignPage'

const { add } = storiesOf('web/campaigns/CreateCampaignPage', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('Page not-dotcom', () => (
    <EnterpriseWebStory>
        {props => (
            <CreateCampaignPage {...props} isSourcegraphDotCom={false} authenticatedUser={{ username: 'alice' }} />
        )}
    </EnterpriseWebStory>
))

add('Page sourcegraph.com', () => (
    <EnterpriseWebStory>
        {props => (
            <CreateCampaignPage {...props} isSourcegraphDotCom={true} authenticatedUser={{ username: 'alice' }} />
        )}
    </EnterpriseWebStory>
))
