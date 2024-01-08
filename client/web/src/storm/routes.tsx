import { uniqBy } from 'lodash'
import type { RouteObject } from 'react-router-dom'

import { routes as legacyRoutes } from '../routes'

import { loader } from './pages/SearchPage/SearchPage.loader'
import { SearchPageWrapper } from './pages/SearchPageWrapper'

export const PagePath = {
    search: '/search',
}

/**
 * How do we migrate routes and keep up with the non-storm version of the app?
 *
 * Proposal:
 *
 * 1. Create storm folder
 * 2. Create routes constant that re-uses existing routes and one by one replaces old ones with new components.
 * 3. "Upgraded" components are moved into the `storm` project.
 *
 */
const stormRoutes: RouteObject[] = [
    {
        path: PagePath.search,
        /**
         * There's still no way to create routes based on query params.
         * Lazy loading of the proper component is handled by `SearchPageWrapper`.
         */
        element: <SearchPageWrapper />,
        loader,
    },
]

export const routes: RouteObject[] = uniqBy([...stormRoutes, ...legacyRoutes], route => route.path)
