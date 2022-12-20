import { act } from '@testing-library/react'
import { createMemoryHistory } from 'history'
import { of } from 'rxjs'

import { renderWithBrandedContext } from '@sourcegraph/wildcard'

import { ProductSubscriptionFieldsOnSubscriptionPage } from '../../../graphql-operations'

import { UserSubscriptionsProductSubscriptionPage } from './UserSubscriptionsProductSubscriptionPage'

describe('UserSubscriptionsProductSubscriptionPage', () => {
    test('renders', () => {
        const history = createMemoryHistory()
        const component = renderWithBrandedContext(
            <UserSubscriptionsProductSubscriptionPage
                user={{ settingsURL: '/u' }}
                match={{
                    isExact: true,
                    params: { subscriptionUUID: '43002662-f627-4550-9af6-d621d2a878de' },
                    path: '/p',
                    url: '/p',
                }}
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
