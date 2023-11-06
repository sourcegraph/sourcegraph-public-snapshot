import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { createFlagMock } from './createFlagMock'
import type { FeatureFlagName } from './featureFlags'
import { withFeatureFlag } from './withFeatureFlag'

describe('withFeatureFlag', () => {
    const trueComponentTestId = 'true-component'
    const falseComponentTestId = 'false-component'
    const TrueComponent = () => <div data-testid={trueComponentTestId}>rendered when flag is true</div>
    const FalseComponent = () => <div data-testid={falseComponentTestId}>rendered when flag is false</div>
    const TEST_FLAG = 'test-flag' as FeatureFlagName
    const Wrapper = withFeatureFlag(TEST_FLAG, TrueComponent, FalseComponent)

    it('renders correctly when flagValue=true', async () => {
        const { findByTestId } = render(
            <MockedTestProvider mocks={[createFlagMock(TEST_FLAG, true)]}>
                <Wrapper />
            </MockedTestProvider>
        )

        expect(await findByTestId(trueComponentTestId)).toBeInTheDocument()
    })

    it('renders correctly when flagValue=false', async () => {
        const { findByTestId } = render(
            <MockedTestProvider mocks={[createFlagMock(TEST_FLAG, false)]}>
                <Wrapper />
            </MockedTestProvider>
        )

        expect(await findByTestId(falseComponentTestId)).toBeInTheDocument()
    })

    it('waits until flag value is resolved', () => {
        const { queryByTestId } = render(
            <MockedTestProvider mocks={[createFlagMock(TEST_FLAG, new Error('Failed to fetch'))]}>
                <Wrapper />
            </MockedTestProvider>
        )

        expect(queryByTestId(falseComponentTestId)).not.toBeInTheDocument()
        expect(queryByTestId(trueComponentTestId)).not.toBeInTheDocument()
    })

    it('renders correctly when flagValue=false and FalseComponent omitted', () => {
        const LocalWrapper = withFeatureFlag('test-flag' as FeatureFlagName, TrueComponent)
        const { findByTestId } = render(
            <MockedTestProvider mocks={[createFlagMock(TEST_FLAG, false)]}>
                <LocalWrapper />
            </MockedTestProvider>
        )

        expect(() => findByTestId(trueComponentTestId)).rejects.toBeTruthy()
    })
})
