import * as H from 'history'
import React from 'react'
import { MemoryRouter } from 'react-router'
import renderer, { act } from 'react-test-renderer'
import { of } from 'rxjs'

import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import { UserSubscriptionsEditProductSubscriptionPage } from './UserSubscriptionsEditProductSubscriptionPage'

jest.mock('mdi-react/ArrowLeftIcon', () => 'ArrowLeftIcon')

jest.mock('./ProductSubscriptionForm', () => ({
    ProductSubscriptionForm: 'ProductSubscriptionForm',
}))

const history = H.createMemoryHistory()
const location = H.createLocation('/')

describe('UserSubscriptionsEditProductSubscriptionPage', () => {
    test('renders', () => {
        const component = renderer.create(
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
        )
        act(() => undefined)
        expect(component.toJSON()).toMatchSnapshot()
    })
})
