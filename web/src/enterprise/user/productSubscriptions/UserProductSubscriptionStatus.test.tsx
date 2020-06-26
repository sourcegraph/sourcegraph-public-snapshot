import React from 'react'
import { UserProductSubscriptionStatus } from './UserProductSubscriptionStatus'
import { mount } from 'enzyme'

jest.mock('mdi-react/KeyIcon', () => 'KeyIcon')
jest.mock('mdi-react/InformationIcon', () => 'InformationIcon')
jest.mock('../../../components/CopyableText', () => ({ CopyableText: 'CopyableText' }))

describe('UserProductSubscriptionStatus', () => {
    test('toggle', () => {
        const component = mount(
            <UserProductSubscriptionStatus
                subscriptionName="L-123"
                productNameWithBrand="P"
                userCount={123}
                expiresAt={23456}
                licenseKey="lk"
            />
        )
        expect(component.children()).toMatchSnapshot('license key hidden')

        // Click "Reveal license key" button.
        component.find('button').simulate('click')
        expect(component.children()).toMatchSnapshot('license key revealed')
    })
})
