import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { CampaignHeader } from './CampaignHeader'
import { subDays } from 'date-fns'

const { add } = storiesOf('web/campaigns/CampaignHeader', module).addDecorator(story => {
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

add('Plain', () => <CampaignHeader />)
add('Full', () => (
    <CampaignHeader
        createdAt={subDays(new Date(), 5).toISOString()}
        creator={{ username: 'alice', url: 'https://test.test/alice' }}
        namespace={{ namespaceName: 'alice', url: 'https://test.test/alice' }}
        name="awesome-campaign"
    />
))
