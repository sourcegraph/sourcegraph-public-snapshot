import { describe, expect, test, jest } from '@jest/globals'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { UserProductSubscriptionStatus } from './UserProductSubscriptionStatus'

jest.mock('mdi-react/KeyIcon', () => () => 'KeyIcon')
jest.mock('mdi-react/InformationIcon', () => () => 'InformationIcon')
jest.mock('../../../components/CopyableText', () => ({ CopyableText: () => 'CopyableText' }))

describe('UserProductSubscriptionStatus', () => {
    test('toggle', () => {
        const { asFragment } = renderWithBrandedContext(
            <UserProductSubscriptionStatus
                subscriptionName="L-123"
                productNameWithBrand="P"
                userCount={123}
                expiresAt={23456}
                licenseKey="lk"
            />
        )
        expect(asFragment()).toMatchSnapshot('license key hidden')

        // Click "Reveal license key" button.
        const button = screen.getByRole('button')
        userEvent.click(button)
        expect(asFragment()).toMatchSnapshot('license key revealed')
    })
})
