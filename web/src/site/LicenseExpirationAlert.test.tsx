import React from 'react'
import { subMonths, addDays } from 'date-fns'
import { LicenseExpirationAlert } from './LicenseExpirationAlert'
import { mount } from 'enzyme'

describe('LicenseExpirationAlert.test.tsx', () => {
    test('expiring soon', () => {
        expect(mount(<LicenseExpirationAlert expiresAt={addDays(new Date(), 3)} daysLeft={3} />)).toMatchSnapshot()
    })

    test('expired', () => {
        expect(mount(<LicenseExpirationAlert expiresAt={subMonths(new Date(), 3)} daysLeft={0} />)).toMatchSnapshot()
    })
})
