import { render } from '@testing-library/react'
import { subMonths, addDays } from 'date-fns'
import React from 'react'

import { LicenseExpirationAlert } from './LicenseExpirationAlert'

describe('LicenseExpirationAlert', () => {
    test('expiring soon', () => {
        expect(
            render(<LicenseExpirationAlert expiresAt={addDays(new Date(), 3)} daysLeft={3} />).asFragment()
        ).toMatchSnapshot()
    })

    test('expired', () => {
        expect(
            render(<LicenseExpirationAlert expiresAt={subMonths(new Date(), 3)} daysLeft={0} />).asFragment()
        ).toMatchSnapshot()
    })
})
