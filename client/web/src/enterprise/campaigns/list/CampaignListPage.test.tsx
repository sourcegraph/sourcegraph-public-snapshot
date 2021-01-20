import * as H from 'history'
import React from 'react'
import { CampaignListPage } from './CampaignListPage'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'
import { of } from 'rxjs'
import { shallow } from 'enzyme'
import { nodes } from './testData'

const history = H.createMemoryHistory()

describe('CampaignListPage', () => {
    for (const totalCount of [0, 1]) {
        test(`renders for siteadmin and totalCount: ${totalCount}`, () => {
            expect(
                shallow(
                    <CampaignListPage
                        history={history}
                        location={history.location}
                        queryCampaigns={() =>
                            of({
                                campaigns: {
                                    totalCount: Object.values(nodes).length,
                                    nodes: Object.values(nodes),
                                    pageInfo: { endCursor: null, hasNextPage: false },
                                },
                                totalCount: Object.values(nodes).length,
                            })
                        }
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                    />
                )
            ).toMatchSnapshot()
        })
        test(`renders for non-siteadmin and totalCount: ${totalCount}`, () => {
            expect(
                shallow(
                    <CampaignListPage
                        history={history}
                        location={history.location}
                        queryCampaigns={() =>
                            of({
                                campaigns: {
                                    totalCount: Object.values(nodes).length,
                                    nodes: Object.values(nodes),
                                    pageInfo: { endCursor: null, hasNextPage: false },
                                },
                                totalCount: Object.values(nodes).length,
                            })
                        }
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                    />
                )
            ).toMatchSnapshot()
        })
    }
})
