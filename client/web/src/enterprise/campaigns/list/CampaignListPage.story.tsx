import { storiesOf } from '@storybook/react'
import React from 'react'
import { CampaignListPage } from './CampaignListPage'
import { nodes } from './testData'
import { of } from 'rxjs'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { useCallback } from '@storybook/addons'

const { add } = storiesOf('web/campaigns/CampaignListPage', module)
    .addDecorator(story => <div className="p-3 container web-content">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const queryCampaigns = () =>
    of({
        campaigns: {
            totalCount: Object.values(nodes).length,
            nodes: Object.values(nodes),
            pageInfo: { endCursor: null, hasNextPage: false },
        },
        totalCount: Object.values(nodes).length,
    })

add('List of campaigns', () => (
    <EnterpriseWebStory>{props => <CampaignListPage {...props} queryCampaigns={queryCampaigns} />}</EnterpriseWebStory>
))

add('No campaigns', () => {
    const queryCampaigns = useCallback(
        () =>
            of({
                campaigns: {
                    totalCount: 0,
                    nodes: [],
                    pageInfo: {
                        endCursor: null,
                        hasNextPage: false,
                    },
                },
                totalCount: 0,
            }),
        []
    )
    return (
        <EnterpriseWebStory>
            {props => <CampaignListPage {...props} queryCampaigns={queryCampaigns} />}
        </EnterpriseWebStory>
    )
})

add('All campaigns tab empty', () => {
    const queryCampaigns = useCallback(
        () =>
            of({
                campaigns: {
                    totalCount: 0,
                    nodes: [],
                    pageInfo: {
                        endCursor: null,
                        hasNextPage: false,
                    },
                },
                totalCount: 0,
            }),
        []
    )
    return (
        <EnterpriseWebStory>
            {props => <CampaignListPage {...props} queryCampaigns={queryCampaigns} openTab="campaigns" />}
        </EnterpriseWebStory>
    )
})
