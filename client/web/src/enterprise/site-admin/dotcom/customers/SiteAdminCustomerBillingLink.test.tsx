import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { SiteAdminCustomerBillingLink } from './SiteAdminCustomerBillingLink'

jest.mock('mdi-react/ExternalLinkIcon', () => 'ExternalLinkIcon')

describe('SiteAdminCustomerBillingLink', () => {
    test('linked billing account', () => {
        expect(
            renderWithBrandedContext(
                <SiteAdminCustomerBillingLink
                    customer={{ id: 'u', urlForSiteAdminBilling: 'https://example.com' }}
                    onDidUpdate={() => undefined}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('no linked billing account', () => {
        expect(
            renderWithBrandedContext(
                <SiteAdminCustomerBillingLink
                    customer={{ id: 'u', urlForSiteAdminBilling: null }}
                    onDidUpdate={() => undefined}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
})
