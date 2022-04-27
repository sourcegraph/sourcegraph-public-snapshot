import { act } from '@testing-library/react'
import { createMemoryHistory } from 'history'
import { of } from 'rxjs'

import * as GQL from '@sourcegraph/shared/src/schema'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { UserSubscriptionsProductSubscriptionPage } from './UserSubscriptionsProductSubscriptionPage'

jest.mock('./ProductSubscriptionHistory', () => ({
    ProductSubscriptionHistory: 'ProductSubscriptionHistory',
}))
describe('UserSubscriptionsProductSubscriptionPage', () => {
    test('renders', () => {
        const history = createMemoryHistory()
        const component = renderWithBrandedContext(
            <UserSubscriptionsProductSubscriptionPage
                user={{ settingsURL: '/u' }}
                match={{ isExact: true, params: { subscriptionUUID: 's' }, path: '/p', url: '/p' }}
                _queryProductSubscription={() =>
                    of<GQL.IProductSubscription>({
                        __typename: 'ProductSubscription',
                    } as GQL.IProductSubscription)
                }
                history={history}
            />,
            { history }
        )
        act(() => undefined)
        expect(component.asFragment()).toMatchSnapshot()
    })
})
