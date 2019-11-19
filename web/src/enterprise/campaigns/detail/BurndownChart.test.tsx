import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignBurndownChart } from './BurndownChart'

// todo(eseliger): remove this once https://github.com/recharts/recharts/pull/1948 is merged
function createNodeMock() {
    const doc = document.implementation.createHTMLDocument()
    return { parentElement: doc.body }
}

describe('CampaignBurndownChart', () => {
    test('renders the chart', () => {
        const history = H.createMemoryHistory({ keyLength: 0 })
        expect(
            renderer
                .create(
                    <CampaignBurndownChart
                        changesetCountsOverTime={[
                            {
                                __typename: 'ChangesetCounts',
                                closed: 1,
                                date: '2019-11-13T23:17:24Z',
                                merged: 1,
                                openApproved: 1,
                                openChangesRequested: 1,
                                openPending: 1,
                                total: 10,
                                open: 8,
                            },
                            {
                                __typename: 'ChangesetCounts',
                                closed: 1,
                                date: '2019-11-14T23:17:24Z',
                                merged: 5,
                                openApproved: 1,
                                openChangesRequested: 1,
                                openPending: 1,
                                total: 10,
                                open: 4,
                            },
                        ]}
                        history={history}
                    />,
                    { createNodeMock }
                )
                .toJSON()
        ).toMatchSnapshot()
    })
    test('renders an alert when data for less than 2 days is present', () => {
        const history = H.createMemoryHistory({ keyLength: 0 })
        expect(
            renderer
                .create(
                    <CampaignBurndownChart
                        changesetCountsOverTime={[
                            {
                                __typename: 'ChangesetCounts',
                                closed: 1,
                                date: '2019-11-13T23:17:24Z',
                                merged: 1,
                                openApproved: 1,
                                openChangesRequested: 1,
                                openPending: 1,
                                total: 10,
                                open: 8,
                            },
                        ]}
                        history={history}
                    />,
                    { createNodeMock }
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
