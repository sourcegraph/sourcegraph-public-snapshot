import { subMonths, addDays } from 'date-fns'
import React from 'react'

import { renderWithRouter } from '@sourcegraph/shared/src/testing/render-with-router'

import { LicenseExpirationAlert } from './LicenseExpirationAlert'

describe('LicenseExpirationAlert', () => {
    test('expiring soon', () => {
        expect(
            renderWithRouter(<LicenseExpirationAlert expiresAt={addDays(new Date(), 3)} daysLeft={3} />).asFragment()
        ).toMatchSnapshot()
    })

    test('expired', () => {
        expect(
            renderWithRouter(<LicenseExpirationAlert expiresAt={subMonths(new Date(), 3)} daysLeft={0} />).asFragment()
        ).toMatchSnapshot()
    })
})
