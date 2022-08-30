import { render } from '@testing-library/react'

import { FeatureFlagName } from './featureFlags'
import { MockedFeatureFlagsProvider } from './FeatureFlagsProvider'
import { withFeatureFlag } from './withFeatureFlag'

describe('withFeatureFlag', () => {
    const trueComponentTestId = 'true-component'
    const falseComponentTestId = 'false-component'
    const TrueComponent = () => <div data-testid={trueComponentTestId}>rendered when flag is true</div>
    const FalseComponent = () => <div data-testid={falseComponentTestId}>rendered when flag is false</div>
    const Wrapper = withFeatureFlag('test-flag' as FeatureFlagName, TrueComponent, FalseComponent)

    it('renders correctly when flagValue=true', async () => {
        const { findByTestId } = render(
            <MockedFeatureFlagsProvider overrides={{ 'test-flag': true } as Partial<Record<FeatureFlagName, boolean>>}>
                <Wrapper />
            </MockedFeatureFlagsProvider>
        )

        expect(await findByTestId(trueComponentTestId)).toBeTruthy()
    })

    it('renders correctly when flagValue=false', async () => {
        const { findByTestId } = render(
            <MockedFeatureFlagsProvider overrides={{ 'test-flag': false } as Partial<Record<FeatureFlagName, boolean>>}>
                <Wrapper />
            </MockedFeatureFlagsProvider>
        )

        expect(await findByTestId(falseComponentTestId)).toBeTruthy()
    })

    it('waits until flag value is resolved', () => {
        const { queryByTestId } = render(
            <MockedFeatureFlagsProvider
                overrides={{ 'test-flag': new Error('Failed to fetch') } as Partial<Record<FeatureFlagName, boolean>>}
            >
                <Wrapper />
            </MockedFeatureFlagsProvider>
        )

        expect(queryByTestId(falseComponentTestId)).toBeFalsy()
        expect(queryByTestId(trueComponentTestId)).toBeFalsy()
    })

    it('renders correctly when flagValue=false and FalseComponent omitted', () => {
        const LocalWrapper = withFeatureFlag('test-flag' as FeatureFlagName, TrueComponent)
        const { findByTestId } = render(
            <MockedFeatureFlagsProvider overrides={{ 'test-flag': false } as Partial<Record<FeatureFlagName, boolean>>}>
                <LocalWrapper />
            </MockedFeatureFlagsProvider>
        )

        expect(() => findByTestId(trueComponentTestId)).rejects.toBeTruthy()
    })
})
