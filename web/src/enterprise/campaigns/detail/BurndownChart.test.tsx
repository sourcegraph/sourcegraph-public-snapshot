import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignBurndownChart } from './BurndownChart'

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
                                __typename: 'ChangesetCounts' as 'ChangesetCounts',
                                closed: 0,
                                date: '2019-11-13T23:17:24Z',
                                merged: 0,
                                openApproved: 0,
                                openChangesRequested: 0,
                                openPending: 1,
                                total: 10,
                                open: 9,
                            },
                            {
                                __typename: 'ChangesetCounts',
                                closed: 0,
                                date: '2019-11-14T23:17:24Z',
                                merged: 0,
                                openApproved: 0,
                                openChangesRequested: 0,
                                openPending: 1,
                                total: 10,
                                open: 9,
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
