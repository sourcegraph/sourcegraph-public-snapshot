import * as H from 'history'
import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { CampaignInfoCard } from './CampaignInfoCard'
import { subMinutes } from 'date-fns'

const { add } = storiesOf('web/campaigns/CampaignInfoCard', module).addDecorator(story => {
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

add('Overview', () => (
    <CampaignInfoCard
        history={H.createMemoryHistory()}
        author={{
            avatarURL: 'http://test.test/asset.png',
            username: 'alice',
        }}
        createdAt={subMinutes(new Date(), 10).toISOString()}
        description="This is an awesome campaign. It will do great things to your codebase."
    />
))
