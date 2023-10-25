import { describe, expect, test } from '@jest/globals'
import { screen } from '@testing-library/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { OnboardingChecklist } from './OnboardingChecklist'
import { completeSiteConfig, incompleteSiteConfig } from './OnboardingChecklist.mocks'

describe('OnboardingChecklist', () => {
    test('render without error', async () => {
        renderWithBrandedContext(
            <MockedTestProvider mocks={[]}>
                <OnboardingChecklist />
            </MockedTestProvider>
        )
        expect(await screen.findByTestId('onboard-loading')).toBeInTheDocument()
        expect(screen.queryByTestId('onboard-dropdown')).not.toBeInTheDocument()
    })

    test('does not render if checklist complete', async () => {
        renderWithBrandedContext(
            <MockedTestProvider mocks={[completeSiteConfig()]}>
                <OnboardingChecklist />
            </MockedTestProvider>
        )
        expect(await screen.findByTestId('onboard-loading')).toBeInTheDocument()
        expect(screen.queryByTestId('onboard-dropdown')).not.toBeInTheDocument()
    })

    test('renders if checklist is missing items', async () => {
        renderWithBrandedContext(
            <MockedTestProvider mocks={[incompleteSiteConfig()]}>
                <OnboardingChecklist />
            </MockedTestProvider>
        )
        expect(await screen.findByTestId('onboard-loading')).toBeInTheDocument()
        expect(await screen.findByText('Setup')).toBeInTheDocument()
    })
})
