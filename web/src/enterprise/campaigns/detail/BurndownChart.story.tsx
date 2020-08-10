import { storiesOf } from '@storybook/react'
import { radios, select } from '@storybook/addon-knobs'
import React from 'react'
import { CampaignBurndownChart } from './BurndownChart'
import { createMemoryHistory } from 'history'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'

const { add } = storiesOf('web/campaigns/BurndownChart', module).addDecorator(story => {
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

add('All states', () => (
    <CampaignBurndownChart
        changesetCountsOverTime={[
            {
                __typename: 'ChangesetCounts' as const,
                date: '2019-11-13T12:00:00Z',
                closed: 0,
                merged: 0,
                openApproved: 0,
                openChangesRequested: 0,
                openPending: 0,
                total: 10,
                open: 10,
            },
            {
                __typename: 'ChangesetCounts' as const,
                date: '2019-11-14T12:00:00Z',
                closed: 0,
                merged: 0,
                openApproved: 0,
                openChangesRequested: 0,
                openPending: 2,
                total: 10,
                open: 10,
            },
            {
                __typename: 'ChangesetCounts' as const,
                date: '2019-11-15T12:00:00Z',
                closed: 0,
                merged: 1,
                openApproved: 1,
                openChangesRequested: 0,
                openPending: 3,
                total: 10,
                open: 8,
            },
            {
                __typename: 'ChangesetCounts' as const,
                date: '2019-11-16T12:00:00Z',
                closed: 1,
                merged: 1,
                openApproved: 1,
                openChangesRequested: 0,
                openPending: 3,
                total: 10,
                open: 7,
            },
            {
                __typename: 'ChangesetCounts' as const,
                date: '2019-11-17T12:00:00Z',
                closed: 1,
                merged: 2,
                openApproved: 1,
                openChangesRequested: 0,
                openPending: 5,
                total: 10,
                open: 6,
            },
            {
                __typename: 'ChangesetCounts' as const,
                date: '2019-11-18T12:00:00Z',
                closed: 2,
                merged: 4,
                openApproved: 0,
                openChangesRequested: 2,
                openPending: 2,
                total: 10,
                open: 4,
            },
            {
                __typename: 'ChangesetCounts' as const,
                date: '2019-11-19T12:00:00Z',
                closed: 2,
                merged: 4,
                openApproved: 0,
                openChangesRequested: 2,
                openPending: 2,
                total: 10,
                open: 4,
            },
            {
                __typename: 'ChangesetCounts' as const,
                date: '2019-11-20T12:00:00Z',
                closed: 2,
                merged: 4,
                openApproved: 0,
                openChangesRequested: 2,
                openPending: 2,
                total: 10,
                open: 4,
            },
            {
                __typename: 'ChangesetCounts' as const,
                date: '2019-11-21T12:00:00Z',
                closed: 2,
                merged: 5,
                openApproved: 3,
                openChangesRequested: 0,
                openPending: 0,
                total: 10,
                open: 3,
            },
            {
                __typename: 'ChangesetCounts' as const,
                date: '2019-11-22T12:00:00Z',
                closed: 2,
                merged: 8,
                openApproved: 0,
                openChangesRequested: 0,
                openPending: 0,
                total: 10,
                open: 0,
            },
        ].slice(0, select('Days of data', [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10], 10))}
        history={createMemoryHistory()}
    />
))
