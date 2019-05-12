import React from 'react'
import { hot } from 'react-hot-loader'

/**
 * Returns a lazy-loaded reference to a React component in another module.
 *
 * This should be used in URL routes and anywhere else that Webpack code splitting can occur, to
 * avoid all referenced components being in the initial bundle.
 *
 * @param moduleFactory Asynchronously imports the component's module; e.g., `() =>
 * import('./MyComponent')`.
 * @param name The export binding name of the component in its module.
 * @param moduleId The Webpack module ID; e.g., `require.resolveWeak('./MyComponent')`.
 */
export const asyncComponent = <P extends {}, K extends string>(
    moduleFactory: () => Promise<{ [k in K]: React.ComponentType<P> }>,
    name: K,
    moduleId: ReturnType<typeof require.resolveWeak>
) =>
    React.lazy(async () => {
        const mod = await moduleFactory()
        if (!module.hot) {
            return { default: mod[name] }
        }

        // Mimic what https://github.com/gaearon/react-hot-loader/blob/master/root.js does, but
        // instead of using the global `module`, use the information passed into `asyncComponent`.
        interface CachedWebpackModule {
            parents: string[]
        }
        const componentModule: CachedWebpackModule = require.cache[moduleId]

        // Remove itself from the cache so it is reloaded when needed.
        delete require.cache[moduleId]

        const hotComponent = hot({ id: moduleId, ...componentModule })(mod[name])
        return { default: hotComponent }
    })
