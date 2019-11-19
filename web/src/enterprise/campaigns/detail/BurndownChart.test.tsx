import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignBurndownChart } from './BurndownChart'

// this is required because recharts internally checks if the chart would be visible.
// Therefore, we need to fake the clientWidth and clientHeight values.
function createNodeMock(element: React.ReactElement<HTMLElement>) {
    if (element.type === 'div') {
        return {
            clientWidth: 500,
            clientHeight: 350,
        }
    }
    return null
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
                                date: '2019-11-13T12:00:00Z',
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
                                date: '2019-11-14T12:00:00Z',
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
                                date: '2019-11-13T12:00:00Z',
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
