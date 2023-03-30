import { act } from '@testing-library/react'
import { of } from 'rxjs'

import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { ProductSubscriptionFieldsOnSubscriptionPage } from '../../../graphql-operations'

import { UserSubscriptionsProductSubscriptionPage } from './UserSubscriptionsProductSubscriptionPage'

describe('UserSubscriptionsProductSubscriptionPage', () => {
    test('renders', () => {
        const component = renderWithBrandedContext(
            <UserSubscriptionsProductSubscriptionPage
                user={{ settingsURL: '/u' }}
                _queryProductSubscription={() =>
                    of<ProductSubscriptionFieldsOnSubscriptionPage>({
                        __typename: 'ProductSubscription',
                    } as ProductSubscriptionFieldsOnSubscriptionPage)
                }
            />,
            { path: '/:subscriptionUUID', route: '/43002662-f627-4550-9af6-d621d2a878de' }
        )
        act(() => undefined)
        expect(component.asFragment()).toMatchSnapshot()
    })
})
