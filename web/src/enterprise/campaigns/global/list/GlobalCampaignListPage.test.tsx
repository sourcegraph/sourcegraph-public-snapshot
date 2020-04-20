import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { GlobalCampaignListPage } from './GlobalCampaignListPage'
import { IUser } from '../../../../../../shared/src/graphql/schema'
import { of } from 'rxjs'

jest.mock('../../../../components/FilteredConnection', () => ({
    FilteredConnection: 'FilteredConnection',
}))

const history = H.createMemoryHistory()

describe('GlobalCampaignListPage', () => {
    for (const totalCount of [0, 1]) {
        test(`renders for siteadmin and totalCount: ${totalCount}`, done => {
            const rendered = renderer.create(
                <GlobalCampaignListPage
                    history={history}
                    location={history.location}
                    authenticatedUser={{ siteAdmin: true } as IUser}
                    queryCampaignsCount={() => of(totalCount)}
                />
            )
            setTimeout(() => {
                expect(rendered.toJSON()).toMatchSnapshot()
                done()
            })
        })
        test(`renders for non-siteadmin and totalCount: ${totalCount}`, done => {
            const rendered = renderer.create(
                <GlobalCampaignListPage
                    history={history}
                    location={history.location}
                    authenticatedUser={{ siteAdmin: false } as IUser}
                    queryCampaignsCount={() => of(totalCount)}
                />
            )
            setTimeout(() => {
                expect(rendered.toJSON()).toMatchSnapshot()
                done()
            })
        })
    }
})
