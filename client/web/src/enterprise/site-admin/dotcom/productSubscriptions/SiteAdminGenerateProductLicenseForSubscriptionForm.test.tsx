import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

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
