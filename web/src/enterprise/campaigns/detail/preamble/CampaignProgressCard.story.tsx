import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../../../../SourcegraphWebApp.scss'
import * as H from 'history'
import { MemoryRouter } from 'react-router'
import { boolean } from '@storybook/addon-knobs'
import { CampaignProgressCard } from './CampaignProgressCard'

const history = H.createMemoryHistory()

const { add } = storiesOf('web/campaigns/CampaignProgressCard', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="theme-light mt-3 container">{story()}</div>
    </>
))

add('Empty', () => (
    <MemoryRouter>
        <CampaignProgressCard
            campaign={{
                id: 'c',
                url: '',
                viewerCanAdminister: boolean('Viewer can administer', true),
            }}
            changesetCounts={{ total: 0, merged: 0, closed: 0, open: 0, unpublished: 0 }}
            history={history}
        />
    </MemoryRouter>
))

add('Partially complete', () => (
    <MemoryRouter>
        <CampaignProgressCard
            campaign={{
                id: 'c',
                url: '',
                viewerCanAdminister: boolean('Viewer can administer', true),
            }}
            changesetCounts={{ total: 107, merged: 23, closed: 8, open: 67, unpublished: 9 }}
            history={history}
        />
    </MemoryRouter>
))

add('Complete', () => (
    <MemoryRouter>
        <CampaignProgressCard
            campaign={{
                id: 'c',
                url: '',
                viewerCanAdminister: boolean('Viewer can administer', true),
            }}
            changesetCounts={{ total: 107, merged: 101, closed: 6, open: 0, unpublished: 0 }}
            history={history}
        />
    </MemoryRouter>
))
