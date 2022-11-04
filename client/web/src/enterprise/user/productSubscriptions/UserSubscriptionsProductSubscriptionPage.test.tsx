import { act } from '@testing-library/react'
import { createMemoryHistory } from 'history'
import { of } from 'rxjs'

import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { ProductSubscriptionFieldsOnSubscriptionPage } from '../../../graphql-operations'

import { UserSubscriptionsProductSubscriptionPage } from './UserSubscriptionsProductSubscriptionPage'

describe('UserSubscriptionsProductSubscriptionPage', () => {
    test('renders', () => {
        const history = createMemoryHistory()
        const component = renderWithBrandedContext(
            <UserSubscriptionsProductSubscriptionPage
                user={{ settingsURL: '/u' }}
                match={{ isExact: true, params: { subscriptionUUID: 's' }, path: '/p', url: '/p' }}
                _queryProductSubscription={() =>
                    of<ProductSubscriptionFieldsOnSubscriptionPage>({
                        __typename: 'ProductSubscription',
                    } as ProductSubscriptionFieldsOnSubscriptionPage)
                }
                history={history}
            />,
            { history }
        )
        act(() => undefined)
        expect(component.asFragment()).toMatchSnapshot()
    })
})
