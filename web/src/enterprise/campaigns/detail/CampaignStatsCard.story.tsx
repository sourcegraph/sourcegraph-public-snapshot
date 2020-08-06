import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { CampaignStatsCard } from './CampaignStatsCard'

const { add } = storiesOf('web/campaigns/CampaignStatsCard', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    return (
        <>
            <Tooltip />
            <style>{webStyles}</style>
            <div className="p-3 container">{story()}</div>
        </>
    )
})

add('All states', () => (
    <CampaignStatsCard
        stats={{
            closed: 10,
            merged: 10,
            open: 10,
            total: 100,
            unpublished: 70,
        }}
    />
))
add('Campaign done', () => (
    <CampaignStatsCard
        stats={{
            closed: 10,
            merged: 90,
            open: 0,
            total: 100,
            unpublished: 0,
        }}
    />
))
