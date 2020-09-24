import { storiesOf } from '@storybook/react'
import React from 'react'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { CampaignStatsCard } from './CampaignStatsCard'

const { add } = storiesOf('web/campaigns/CampaignStatsCard', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('All states', () => (
    <EnterpriseWebStory>
        {props => (
            <CampaignStatsCard
                {...props}
                stats={{
                    closed: 10,
                    merged: 10,
                    open: 10,
                    total: 100,
                    unpublished: 70,
                }}
                closedAt={null}
            />
        )}
    </EnterpriseWebStory>
))
add('Campaign closed', () => (
    <EnterpriseWebStory>
        {props => (
            <CampaignStatsCard
                {...props}
                stats={{
                    closed: 10,
                    merged: 10,
                    open: 10,
                    total: 100,
                    unpublished: 70,
                }}
                closedAt={new Date().toISOString()}
            />
        )}
    </EnterpriseWebStory>
))
add('Campaign done', () => (
    <EnterpriseWebStory>
        {props => (
            <CampaignStatsCard
                {...props}
                stats={{
                    closed: 10,
                    merged: 90,
                    open: 0,
                    total: 100,
                    unpublished: 0,
                }}
                closedAt={null}
            />
        )}
    </EnterpriseWebStory>
))
