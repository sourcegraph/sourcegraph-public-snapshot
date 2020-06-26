import React from 'react'
import * as H from 'history'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { UserSubscriptionsEditProductSubscriptionPage } from './UserSubscriptionsEditProductSubscriptionPage'
import { of } from 'rxjs'
import { MemoryRouter } from 'react-router'
import { mount } from 'enzyme'

jest.mock('mdi-react/ArrowLeftIcon', () => 'ArrowLeftIcon')

jest.mock('./ProductSubscriptionForm', () => ({
    ProductSubscriptionForm: 'ProductSubscriptionForm',
}))

jest.mock('../../../tracking/eventLogger', () => ({
    eventLogger: { logViewEvent: () => undefined },
}))

const history = H.createMemoryHistory()
const location = H.createLocation('/')

describe('UserSubscriptionsEditProductSubscriptionPage', () => {
    test('renders', () => {
        expect(
            mount(
                <MemoryRouter>
                    <UserSubscriptionsEditProductSubscriptionPage
                        user={{ id: 'u' }}
                        match={{ isExact: true, params: { subscriptionUUID: 's' }, path: '/p', url: '/p' }}
                        history={history}
                        location={location}
                        isLightTheme={true}
                        _queryProductSubscription={() =>
                            of({
                                id: 's',
                                name: 'L-123',
                                invoiceItem: {
                                    __typename: 'ProductSubscriptionInvoiceItem',
                                    userCount: 123,
                                    expiresAt: '2020-01-01',
                                    plan: { __typename: 'ProductPlan', billingPlanID: 'bp' } as GQL.IProductPlan,
                                },
                                url: '/s',
                            })
                        }
                    />
                </MemoryRouter>
            ).children()
        ).toMatchSnapshot()
    })
})
