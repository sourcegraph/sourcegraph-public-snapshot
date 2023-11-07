import { noop } from 'lodash'
import { afterEach, beforeEach, describe, expect, test } from 'vitest'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { SiteAdminGenerateProductLicenseForSubscriptionForm } from './SiteAdminGenerateProductLicenseForSubscriptionForm'
import { mockLicenseContext, mockLicense } from './testUtils'

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
                        latestLicense={mockLicense}
                        onCancel={noop}
                        subscriptionAccount="foo"
                        onGenerate={noop}
                    />
                </MockedTestProvider>
            ).baseElement
        ).toMatchSnapshot()
    })
})
