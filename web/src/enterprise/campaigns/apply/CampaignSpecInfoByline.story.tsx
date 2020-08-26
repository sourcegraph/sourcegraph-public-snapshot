import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { CampaignSpecInfoByline } from './CampaignSpecInfoByline'
import { subDays } from 'date-fns'

const { add } = storiesOf('web/campaigns/apply/CampaignSpecInfoByline', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    return (
        <>
            <Tooltip />
            <style>{webStyles}</style>
            <div className="p-3 container web-content">{story()}</div>
        </>
    )
})

add('Default', () => (
    <CampaignSpecInfoByline
        createdAt={subDays(new Date(), 3).toISOString()}
        creator={{ url: 'http://test.test/alice', username: 'alice' }}
    />
))
