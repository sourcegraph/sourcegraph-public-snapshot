import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import { CampaignsBetaFeedbackAlert } from './CampaignsBetaFeedbackAlert'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'

const { add } = storiesOf('web/campaigns/CampaignsBetaFeedbackAlert', module).addDecorator(story => {
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

add('Alert box', () => {
    // Pass in a random string, so trying out the dismiss buttton doesn't render the storybook useless.
    const randomString = String(Date.now())
    return <CampaignsBetaFeedbackAlert partialStorageKey={randomString} />
})
