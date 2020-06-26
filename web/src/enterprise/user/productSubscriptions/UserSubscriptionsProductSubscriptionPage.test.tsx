import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { UserSubscriptionsProductSubscriptionPage } from './UserSubscriptionsProductSubscriptionPage'
import { of } from 'rxjs'
import { MemoryRouter } from 'react-router'
import { createMemoryHistory } from 'history'
import { mount } from 'enzyme'

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

jest.mock('../../../tracking/eventLogger', () => ({
    eventLogger: { logViewEvent: () => undefined },
}))

describe('UserSubscriptionsProductSubscriptionPage', () => {
    test('renders', () => {
        expect(
            mount(
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
            ).children()
        ).toMatchSnapshot()
    })
})
