import { render } from '@testing-library/react'
import React from 'react'

import { SiteAdminGenerateProductLicenseForSubscriptionForm } from './SiteAdminGenerateProductLicenseForSubscriptionForm'

describe('SiteAdminGenerateProductLicenseForSubscriptionForm', () => {
    test('renders', () => {
        expect(
            render(
                <SiteAdminGenerateProductLicenseForSubscriptionForm subscriptionID="s" onGenerate={() => undefined} />
            ).asFragment()
        ).toMatchSnapshot()
    })
})
