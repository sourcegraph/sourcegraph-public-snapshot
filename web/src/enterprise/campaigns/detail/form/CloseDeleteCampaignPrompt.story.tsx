import { storiesOf } from '@storybook/react'
import { radios, boolean } from '@storybook/addon-knobs'
import React from 'react'
import { CloseDeleteCampaignPrompt } from './CloseDeleteCampaignPrompt'
import webStyles from '../../../../enterprise.scss'
import { Tooltip } from '../../../../components/tooltip/Tooltip'
import { action } from '@storybook/addon-actions'

const { add } = storiesOf('web/campaigns/CloseDeleteCampaignPrompt', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    return (
        <>
            <Tooltip />
            <style>{webStyles}</style>
            <div className="p-3 container">
                <div className="position-relative d-flex justify-content-end">{story()}</div>
            </div>
        </>
    )
})

add('Close button', () => (
    <CloseDeleteCampaignPrompt
        disabled={boolean('disabled', false)}
        disabledTooltip="Cannot close while campaign is being created"
        message={
            <p>
                Close campaign <strong>awesome-campaign</strong>?
            </p>
        }
        buttonText="Close"
        onButtonClick={action('Button clicked')}
        buttonClassName="btn-secondary"
        initiallyOpen={true}
    />
))
