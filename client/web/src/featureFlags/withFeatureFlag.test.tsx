import { render } from '@testing-library/react'

import { FeatureFlagName } from './featureFlags'
import { MockedFeatureFlagsProvider } from './FeatureFlagsProvider'
import { withFeatureFlag } from './withFeatureFlag'

describe('withFeatureFlag', () => {
    const TrueComponent = () => <div data-testid="true-component">rendered when flag is true</div>
    const FalseComponent = () => <div data-testid="false-component">rendered when flag is false</div>
    const Wrapper = withFeatureFlag('test-flag' as FeatureFlagName, TrueComponent, FalseComponent)

    it('renders correctly when flagValue=true', () => {
        const { getByTestId } = render(
            <MockedFeatureFlagsProvider overrides={new Map([['test-flag' as FeatureFlagName, true]])}>
                <Wrapper />
            </MockedFeatureFlagsProvider>
        )

        expect(getByTestId('true-component')).toBeTruthy()
    })

    it('renders correctly when flagValue=false', () => {
        const { getByTestId } = render(
            <MockedFeatureFlagsProvider overrides={new Map([['test-flag' as FeatureFlagName, false]])}>
                <Wrapper />
            </MockedFeatureFlagsProvider>
        )

        expect(getByTestId('false-component')).toBeTruthy()
    })

    it('waits until flag value is resolved', () => {
        const { queryByTestId } = render(
            <MockedFeatureFlagsProvider
                overrides={new Map([['test-flag' as FeatureFlagName, new Error('Failed to fetch')]])}
            >
                <Wrapper />
            </MockedFeatureFlagsProvider>
        )

        expect(queryByTestId('false-component')).toBeFalsy()
        expect(queryByTestId('true-component')).toBeFalsy()
    })

    it('renders correctly when flagValue=false and FalseComponent omitted', () => {
        const LocalWrapper = withFeatureFlag('test-flag' as FeatureFlagName, TrueComponent)
        const { container } = render(
            <MockedFeatureFlagsProvider overrides={new Map([['test-flag' as FeatureFlagName, false]])}>
                <LocalWrapper />
            </MockedFeatureFlagsProvider>
        )

        expect(container.innerHTML).toBeFalsy()
    })
})
