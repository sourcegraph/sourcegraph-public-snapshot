import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import { CampaignListPage } from './CampaignListPage'
import { createMemoryHistory } from 'history'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'
import { nodes } from './CampaignNode.story'
import { of } from 'rxjs'

const { add } = storiesOf('web/campaigns/CampaignListPage', module).addDecorator(story => {
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

add('List of campaigns', () => {
    const history = createMemoryHistory()
    return (
        <CampaignListPage
            queryCampaigns={() =>
                of({
                    totalCount: Object.values(nodes).length,
                    nodes: Object.values(nodes),
                    pageInfo: { endCursor: null, hasNextPage: false },
                })
            }
            telemetryService={NOOP_TELEMETRY_SERVICE}
            history={history}
            location={history.location}
        />
    )
})
