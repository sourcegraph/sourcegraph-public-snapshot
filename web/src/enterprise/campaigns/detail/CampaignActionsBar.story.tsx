import { storiesOf } from '@storybook/react'
import { radios, boolean } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { CampaignActionsBar } from './CampaignActionsBar'

const { add } = storiesOf('web/campaigns/CampaignActionsBar', module).addDecorator(story => {
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

add('Bar', () => (
    <CampaignActionsBar
        campaign={{
            name: 'Awesome campaign',
            closedAt: boolean('Closed', false) ? new Date().toISOString() : null,
            viewerCanAdminister: boolean('viewerCanAdminister', false),
            changesets: {
                totalCount: 107,
                stats: {
                    total: 107,
                    closed: 10,
                    merged: 20,
                },
            },
        }}
    />
))
