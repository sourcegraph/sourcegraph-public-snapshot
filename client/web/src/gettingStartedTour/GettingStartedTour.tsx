import React from 'react'

import { ErrorBoundary } from '../components/ErrorBoundary'

import { GettingStartedTourManager, GettingStartedTourManagerProps } from './GettingStartedTourManager'

export const GettingStartedTour: React.FunctionComponent<GettingStartedTourManagerProps> = props => (
    <ErrorBoundary
        location={null}
        render={error => (
            <div>
                Getting Started Tour: Something went wrong :(.
                <pre>{JSON.stringify(error)}</pre>
            </div>
        )}
    >
        <GettingStartedTourManager {...props} />
    </ErrorBoundary>
)
