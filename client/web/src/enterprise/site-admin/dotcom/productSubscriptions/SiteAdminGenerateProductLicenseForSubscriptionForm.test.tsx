import { renderWithBrandedContext } from '@sourcegraph/wildcard'

import { SiteAdminGenerateProductLicenseForSubscriptionForm } from './SiteAdminGenerateProductLicenseForSubscriptionForm'

describe('SiteAdminGenerateProductLicenseForSubscriptionForm', () => {
    test('renders', () => {
        expect(
            renderWithBrandedContext(
                <SiteAdminGenerateProductLicenseForSubscriptionForm subscriptionID="s" onGenerate={() => undefined} />
            ).asFragment()
        ).toMatchSnapshot()
    })
})
