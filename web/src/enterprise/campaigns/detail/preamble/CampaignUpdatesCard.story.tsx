import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../../../../SourcegraphWebApp.scss'
import * as H from 'history'
import { MemoryRouter } from 'react-router'
import { boolean } from '@storybook/addon-knobs'
import { CampaignUpdatesCard } from './CampaignUpdatesCard'

const history = H.createMemoryHistory()

const { add } = storiesOf('web/campaigns/CampaignUpdatesCard', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="theme-light mt-3 container">{story()}</div>
    </>
))

add('Empty campaign', () => (
    <MemoryRouter>
        <CampaignUpdatesCard
            campaign={{
                id: 'c',
                url: '',
                patchSetter: null,
                patchesSetAt: null,
                changesets: { totalCount: 0 },
                viewerCanAdminister: boolean('Viewer can administer', true),
                closedAt: boolean('Campaign closed', false) ? null : '2020-01-01',
            }}
            history={history}
        />
    </MemoryRouter>
))

add('With patches', () => (
    <MemoryRouter>
        <CampaignUpdatesCard
            campaign={{
                id: 'c',
                url: '',
                patchSetter: { username: 'alice', url: '/users/alice' },
                patchesSetAt: '2020-01-01',
                changesets: { totalCount: 3 },
                viewerCanAdminister: boolean('Viewer can administer', true),
                closedAt: boolean('Campaign closed', false) ? null : '2020-01-01',
            }}
            history={history}
        />
    </MemoryRouter>
))

add('With tracked changesets only', () => (
    <MemoryRouter>
        <CampaignUpdatesCard
            campaign={{
                id: 'c',
                url: '',
                patchSetter: null,
                patchesSetAt: null,
                changesets: { totalCount: 3 },
                viewerCanAdminister: boolean('Viewer can administer', true),
                closedAt: boolean('Campaign closed', false) ? null : '2020-01-01',
            }}
            history={history}
        />
    </MemoryRouter>
))
