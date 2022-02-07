import { render, act } from '@testing-library/react'
import * as H from 'history'
import React from 'react'
import { MemoryRouter } from 'react-router'
import { of } from 'rxjs'

import * as GQL from '@sourcegraph/shared/src/schema'

import { UserSubscriptionsEditProductSubscriptionPage } from './UserSubscriptionsEditProductSubscriptionPage'

jest.mock('mdi-react/ArrowLeftIcon', () => 'ArrowLeftIcon')

const history = H.createMemoryHistory()
const location = H.createLocation('/')

describe('UserSubscriptionsEditProductSubscriptionPage', () => {
    test('renders', () => {
        const component = render(
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
        expect(component.asFragment()).toMatchSnapshot()
    })
})
