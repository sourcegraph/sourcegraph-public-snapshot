import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { CampaignInfoByline } from './CampaignInfoByline'
import { subDays } from 'date-fns'

const { add } = storiesOf('web/campaigns/CampaignInfoByline', module).addDecorator(story => {
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
    <CampaignInfoByline
        createdAt={subDays(new Date(), 3).toISOString()}
        initialApplier={{ url: 'http://test.test/alice', username: 'alice' }}
        lastAppliedAt={subDays(new Date(), 1).toISOString()}
        lastApplier={{ url: 'http://test.test/bob', username: 'bob' }}
    />
))
