import React from 'react'

import { ErrorBoundary } from '../../components/ErrorBoundary'

export const withErrorBoundary = <P extends object>(
    Component: React.ComponentType<P>,
    errorMessage: React.ReactNode = 'Something went wrong :(.'
) => (props: P) => (
    <ErrorBoundary
        location={null}
        render={error => (
            <div>
                {errorMessage}
                <pre>{JSON.stringify(error)}</pre>
            </div>
        )}
    >
        <Component {...props} />
    </ErrorBoundary>
)
