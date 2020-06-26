import React from 'react'
import { SiteAdminProductSubscriptionBillingLink } from './SiteAdminProductSubscriptionBillingLink'
import { mount } from 'enzyme'

jest.mock('mdi-react/ExternalLinkIcon', () => 'ExternalLinkIcon')

describe('SiteAdminProductSubscriptionBillingLink', () => {
    test('linked billing', () => {
        expect(
            mount(
                <SiteAdminProductSubscriptionBillingLink
                    productSubscription={{ id: 'u', urlForSiteAdminBilling: 'https://example.com' }}
                    onDidUpdate={() => undefined}
                />
            ).children()
        ).toMatchSnapshot()
    })

    test('no linked billing', () => {
        expect(
            mount(
                <SiteAdminProductSubscriptionBillingLink
                    productSubscription={{ id: 'u', urlForSiteAdminBilling: null }}
                    onDidUpdate={() => undefined}
                />
            ).children()
        ).toMatchSnapshot()
    })
})
