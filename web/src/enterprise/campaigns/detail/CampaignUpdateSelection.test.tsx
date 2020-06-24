import React from 'react'
import * as H from 'history'
import { CampaignUpdateSelection } from './CampaignUpdateSelection'
import { shallow } from 'enzyme'

describe('CampaignUpdateSelection', () => {
    test('renders with URL param', () => {
        const history = H.createMemoryHistory({ initialEntries: ['/campaigns/update?patchSet=123'] })
        expect(
            shallow(<CampaignUpdateSelection history={history} location={history.location} className="abc" />)
        ).toMatchSnapshot()
    })
    test('redirects without URL param', () => {
        const history = H.createMemoryHistory({ initialEntries: ['/campaigns/update'] })
        expect(
            shallow(<CampaignUpdateSelection history={history} location={history.location} className="abc" />)
        ).toMatchSnapshot()
    })
})
