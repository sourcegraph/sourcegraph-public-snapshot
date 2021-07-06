import { createMemoryHistory } from 'history'
import React from 'react'
import { MemoryRouter } from 'react-router'
import renderer, { act } from 'react-test-renderer'
import { of } from 'rxjs'

import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import { UserSubscriptionsProductSubscriptionPage } from './UserSubscriptionsProductSubscriptionPage'

jest.mock('./BackToAllSubscriptionsLink', () => ({
    BackToAllSubscriptionsLink: 'BackToAllSubscriptionsLink',
}))

jest.mock('./UserProductSubscriptionStatus', () => ({
    UserProductSubscriptionStatus: 'UserProductSubscriptionStatus',
}))

jest.mock('./ProductSubscriptionBilling', () => ({
    ProductSubscriptionBilling: 'ProductSubscriptionBilling',
}))

jest.mock('./ProductSubscriptionHistory', () => ({
    ProductSubscriptionHistory: 'ProductSubscriptionHistory',
}))

describe('UserSubscriptionsProductSubscriptionPage', () => {
    test('renders', () => {
        const component = renderer.create(
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
        expect(component.toJSON()).toMatchSnapshot()
    })
})
