import React, { type Attributes, type PropsWithChildren, type PropsWithRef } from 'react'

import { TelemetryV2Props } from '../telemetry'

/**
 * Returns a lazy-loaded reference to a React component in another module.
 *
 * This should be used in URL routes and anywhere else that code splitting can occur, to
 * avoid all referenced components being in the initial bundle.
 *
 * @param componentFactory Asynchronously imports the component's module; e.g., `() =>
 * import('./MyComponent')`.
 * @param name The export binding name of the component in its module.
 */
export const lazyComponent = <P extends {}, K extends string>(
    componentFactory: () => Promise<{ [k in K]: React.ComponentType<React.PropsWithChildren<P>> }>,
    name: K
): React.FunctionComponent<
    React.PropsWithChildren<PropsWithRef<PropsWithChildren<P>> & Attributes & TelemetryV2Props>
> => {
    // Force returning a React.FunctionComponent-like so our result is callable (because it's used
    // in <Route render={...} /> elements where it is expected to be callable).
    const LazyComponent = React.lazy(async () => {
        const component: React.ComponentType<React.PropsWithChildren<P>> = (await componentFactory())[name]
        return { default: component }
    })
    return props => <LazyComponent {...props} />
}
