import React from 'react'
import { SiteAdminGenerateProductLicenseForSubscriptionForm } from './SiteAdminGenerateProductLicenseForSubscriptionForm'
import { createMemoryHistory } from 'history'
import { mount } from 'enzyme'

describe('SiteAdminGenerateProductLicenseForSubscriptionForm', () => {
    test('renders', () => {
        expect(
            mount(
                <SiteAdminGenerateProductLicenseForSubscriptionForm
                    subscriptionID="s"
                    onGenerate={() => undefined}
                    history={createMemoryHistory()}
                />
            ).children()
        ).toMatchSnapshot()
    })
})
