import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { ExtensionsAreaRoute } from '../../extensions/ExtensionsArea'
import { extensionsAreaRoutes } from '../../extensions/routes'

export const enterpriseExtensionsAreaRoutes: readonly ExtensionsAreaRoute[] = [
    extensionsAreaRoutes[0],
    {
        path: '/registry',
        render: lazyComponent(() => import('./registry/RegistryArea'), 'RegistryArea'),
    },
    ...extensionsAreaRoutes.slice(1),
]
