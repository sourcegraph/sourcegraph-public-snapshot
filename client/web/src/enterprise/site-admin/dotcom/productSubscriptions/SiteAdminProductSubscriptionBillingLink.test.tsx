import { render } from '@testing-library/react'
import React from 'react'

import { SiteAdminProductSubscriptionBillingLink } from './SiteAdminProductSubscriptionBillingLink'

jest.mock('mdi-react/ExternalLinkIcon', () => 'ExternalLinkIcon')

describe('SiteAdminProductSubscriptionBillingLink', () => {
    test('linked billing', () => {
        expect(
            render(
                <SiteAdminProductSubscriptionBillingLink
                    productSubscription={{ id: 'u', urlForSiteAdminBilling: 'https://example.com' }}
                    onDidUpdate={() => undefined}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('no linked billing', () => {
        expect(
            render(
                <SiteAdminProductSubscriptionBillingLink
                    productSubscription={{ id: 'u', urlForSiteAdminBilling: null }}
                    onDidUpdate={() => undefined}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
})
