import { render, act } from '@testing-library/react'
import { createMemoryHistory } from 'history'
import React from 'react'
import { MemoryRouter } from 'react-router'
import { of } from 'rxjs'

import * as GQL from '@sourcegraph/shared/src/schema'

import { UserSubscriptionsProductSubscriptionPage } from './UserSubscriptionsProductSubscriptionPage'

jest.mock('./ProductSubscriptionHistory', () => ({
    ProductSubscriptionHistory: 'ProductSubscriptionHistory',
}))
describe('UserSubscriptionsProductSubscriptionPage', () => {
    test('renders', () => {
        const component = render(
            <MemoryRouter>
                <UserSubscriptionsProductSubscriptionPage
                    user={{ settingsURL: '/u' }}
                    match={{ isExact: true, params: { subscriptionUUID: 's' }, path: '/p', url: '/p' }}
                    _queryProductSubscription={() =>
                        of<GQL.IProductSubscription>({
                            __typename: 'ProductSubscription',
                        } as GQL.IProductSubscription)
                    }
                    history={createMemoryHistory()}
                />
            </MemoryRouter>
        )
        act(() => undefined)
        expect(component.asFragment()).toMatchSnapshot()
    })
})
