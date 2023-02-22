import { createPlugin, createRoutableExtension } from '@backstage/core-plugin-api'

import { rootRouteRef } from './routes'

export const sourcegraphPlugin = createPlugin({
    id: 'sourcegraph',
    routes: {
        root: rootRouteRef,
    },
})

export const SourcegraphPage = sourcegraphPlugin.provide(
    createRoutableExtension({
        name: 'SourcegraphPage',
        component: () => import('./components/ExampleComponent').then(m => m.ExampleComponent),
        mountPoint: rootRouteRef,
    })
)
