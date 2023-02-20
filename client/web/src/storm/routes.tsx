import { uniqBy } from 'lodash'

import { routes as legacyRoutes } from '../routes'

import { SearchPageWrapper } from './pages/SearchPageWrapper'
import { loader } from './pages/SearchPage/SearchPage.loader'

export const PagePath = {
    search: '/search',
}

/**
 * It would be great to lazy-load `element` and `loader` in parallel.
 * The last option in this article describes how to do it and avoid CLS:
 * https://dev.to/infoxicator/data-routers-code-splitting-23no
 *
 * But it's not clear how to colocate GraphQL fragments with components.
 * For codesplitting purposes the only options would be to have them in separate files.
 *
 * - Page.tsx
 * - Page.loader.ts
 *   - ComponentA.tsx
 *   - ComponentA.loader.ts
 *   - ComponentB.tsx
 *   - ComponentB.loader.ts
 *
 * The in page we can still have a separate chunk for data-loader which uses nested loader files.

 * But remix team went with a simpler approach where by default the rely on
 * loader being in the same chunk as the route component:
 * https://github.com/remix-run/react-router/issues/9884
 *
 * Initially we can go with the simpler approach to and experiment with separate loader/component
 * chunks later.Thoughts?
 *
 * TODO: start using filename routing convention.
 * File based routing?
 *
 * ------------------------------------------
 *
 * How do we migrate routes and keep up with the non-storm version of the app?
 *
 * Proposal:
 *
 * 1. Create storm folder for OSS and enterprise version.
 * 2. Create routes constant that re-uses existing routes and one by one replaces old ones with new components.
 * 3. "Upgraded" component are moved into the `storm` project
 *
 * ------------------------------
 *
 * Code generation?
 * TODO: create an issue to wrap storm routes in <Strict />
 */
export const routes = uniqBy(
    [
        {
            path: PagePath.search,
            /**
             * There's still no way to create routes based on query params.
             * Lazy loading of the proper component is handled by `SearchPageWrapper`.
             */
            render: props => <SearchPageWrapper {...props} />,
            loader,
        },
        ...legacyRoutes,
    ],
    route => route.path
)
