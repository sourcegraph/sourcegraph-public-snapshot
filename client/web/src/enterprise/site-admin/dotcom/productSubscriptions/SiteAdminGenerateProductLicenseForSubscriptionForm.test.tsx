import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { SiteAdminGenerateProductLicenseForSubscriptionForm } from './SiteAdminGenerateProductLicenseForSubscriptionForm'

describe('SiteAdminGenerateProductLicenseForSubscriptionForm', () => {
    const origContext = window.context
    beforeEach(() => {
        window.context = {
            licenseInfo: {
                knownLicenseTags: ['plan:enterprise-1', 'batch-changes', 'true-up', 'trial'],
            },
        } as any
    })
    afterEach(() => {
        window.context = origContext
    })

    test('renders', () => {
        expect(
            renderWithBrandedContext(
                <MockedTestProvider mocks={[]}>
                    <SiteAdminGenerateProductLicenseForSubscriptionForm
                        subscriptionID="s"
                        subscriptionAccount="foo"
                        onGenerate={() => undefined}
                    />
                </MockedTestProvider>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
