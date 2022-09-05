import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

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
