import { describe, expect, test } from '@jest/globals'
import { subMonths, addDays } from 'date-fns'

import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { LicenseExpirationAlert } from './LicenseExpirationAlert'

describe('LicenseExpirationAlert', () => {
    test('expiring soon', () => {
        expect(
            renderWithBrandedContext(
                <LicenseExpirationAlert expiresAt={addDays(new Date(), 3)} daysLeft={3} />
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('expired', () => {
        expect(
            renderWithBrandedContext(
                <LicenseExpirationAlert expiresAt={subMonths(new Date(), 3)} daysLeft={0} />
            ).asFragment()
        ).toMatchSnapshot()
    })
})
