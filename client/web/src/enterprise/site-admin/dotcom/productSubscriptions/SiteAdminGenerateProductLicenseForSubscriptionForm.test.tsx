import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { SiteAdminGenerateProductLicenseForSubscriptionForm } from './SiteAdminGenerateProductLicenseForSubscriptionForm'
import { mockLicenseContext } from './testUtils'

describe('SiteAdminGenerateProductLicenseForSubscriptionForm', () => {
    const origContext = window.context
    beforeEach(() => {
        window.context = mockLicenseContext
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
