import React from 'react'

import { ErrorBoundary } from '../components/ErrorBoundary'

import { OnboardingTourManager, OnboardingTourManagerProps } from './OnboardingTourManager'

export const OnboardingTour: React.FunctionComponent<OnboardingTourManagerProps> = props => (
    <ErrorBoundary
        location={null}
        render={error => (
            <div>
                Onboarding Tour: Something went wrong :(.
                <pre>{JSON.stringify(error)}</pre>
            </div>
        )}
    >
        <OnboardingTourManager {...props} />
    </ErrorBoundary>
)
