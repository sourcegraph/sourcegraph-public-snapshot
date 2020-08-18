import * as H from 'history'
import { storiesOf } from '@storybook/react'
import { radios, boolean } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { CreateUpdateCampaignAlert } from './CreateUpdateCampaignAlert'
import { noop } from 'lodash'

const { add } = storiesOf('web/campaigns/apply/CreateUpdateCampaignAlert', module).addDecorator(story => {
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

add('Create', () => {
    const history = H.createMemoryHistory()
    return (
        <CreateUpdateCampaignAlert
            specID="123"
            campaign={null}
            history={history}
            isLoading={false}
            setIsLoading={noop}
            viewerCanAdminister={boolean('viewerCanAdminister', true)}
        />
    )
})
add('Update', () => {
    const history = H.createMemoryHistory()
    return (
        <CreateUpdateCampaignAlert
            specID="123"
            campaign={{ id: '123', name: 'awesome-campaign', url: 'http://test.test/awesome' }}
            history={history}
            isLoading={false}
            setIsLoading={noop}
            viewerCanAdminister={boolean('viewerCanAdminister', true)}
        />
    )
})
