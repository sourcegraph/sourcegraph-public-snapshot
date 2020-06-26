import React from 'react'
import { SiteAdminCustomerBillingLink } from './SiteAdminCustomerBillingLink'
import { mount } from 'enzyme'

jest.mock('mdi-react/ExternalLinkIcon', () => 'ExternalLinkIcon')

describe('SiteAdminCustomerBillingLink', () => {
    test('linked billing account', () => {
        expect(
            mount(
                <SiteAdminCustomerBillingLink
                    customer={{ id: 'u', urlForSiteAdminBilling: 'https://example.com' }}
                    onDidUpdate={() => undefined}
                />
            ).children()
        ).toMatchSnapshot()
    })

    test('no linked billing account', () => {
        expect(
            mount(
                <SiteAdminCustomerBillingLink
                    customer={{ id: 'u', urlForSiteAdminBilling: null }}
                    onDidUpdate={() => undefined}
                />
            ).children()
        ).toMatchSnapshot()
    })
})
