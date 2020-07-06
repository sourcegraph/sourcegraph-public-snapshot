import * as H from 'history'
import React from 'react'
import { GlobalCampaignListPage } from './GlobalCampaignListPage'
import { IUser } from '../../../../../../shared/src/graphql/schema'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../../shared/src/telemetry/telemetryService'
import { of } from 'rxjs'
import { mount } from 'enzyme'

jest.mock('../../../../components/FilteredConnection', () => ({
    FilteredConnection: 'FilteredConnection',
}))

const history = H.createMemoryHistory()

describe('GlobalCampaignListPage', () => {
    for (const totalCount of [0, 1]) {
        test(`renders for siteadmin and totalCount: ${totalCount}`, () => {
            expect(
                mount(
                    <GlobalCampaignListPage
                        history={history}
                        location={history.location}
                        authenticatedUser={{ siteAdmin: true } as IUser}
                        queryCampaignsCount={() => of(totalCount)}
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                    />
                )
            ).toMatchSnapshot()
        })
        test(`renders for non-siteadmin and totalCount: ${totalCount}`, () => {
            expect(
                mount(
                    <GlobalCampaignListPage
                        history={history}
                        location={history.location}
                        authenticatedUser={{ siteAdmin: false } as IUser}
                        queryCampaignsCount={() => of(totalCount)}
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                    />
                )
            ).toMatchSnapshot()
        })
    }
})
