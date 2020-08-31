import { storiesOf } from '@storybook/react'
import React from 'react'
import { WebStory } from '../../../components/WebStory'
import webStyles from '../../../enterprise.scss'
import { CampaignStatsCard } from './CampaignStatsCard'

const { add } = storiesOf('web/campaigns/CampaignStatsCard', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('All states', () => (
    <WebStory webStyles={webStyles}>
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
    </WebStory>
))
add('Campaign closed', () => (
    <WebStory webStyles={webStyles}>
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
    </WebStory>
))
add('Campaign done', () => (
    <WebStory webStyles={webStyles}>
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
    </WebStory>
))
