import { storiesOf } from '@storybook/react'
import React from 'react'
import { CampaignListPage } from './CampaignListPage'
import { nodes } from './CampaignNode.story'
import { of } from 'rxjs'
import { EnterpriseWebStory } from '../../../components/WebStory'

const { add } = storiesOf('web/campaigns/CampaignListPage', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

const queryCampaigns = () =>
    of({
        totalCount: Object.values(nodes).length,
        nodes: Object.values(nodes),
        pageInfo: { endCursor: null, hasNextPage: false },
    })

add('List of campaigns', () => (
    <EnterpriseWebStory>{props => <CampaignListPage {...props} queryCampaigns={queryCampaigns} />}</EnterpriseWebStory>
))
