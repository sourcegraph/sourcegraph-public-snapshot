import React from 'react'

import { ErrorBoundary } from '../../components/ErrorBoundary'

/**
 * HOC. Wraps a given component w/ ErrorBoundary component.
 *
 * @param Component a component to render
 * @param errorMessage a custom message to show when error is catched
 */
export const withErrorBoundary =
    <P extends object>(
        Component: React.ComponentType<React.PropsWithChildren<P>>,
        errorMessage: React.ReactNode = 'Something went wrong :(.'
    ) =>
    (props: P): React.ReactElement =>
        (
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
