import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { ExtensionAreaRoute } from '../../../extensions/extension/ExtensionArea'
import { extensionAreaRoutes } from '../../../extensions/extension/routes'

export const enterpriseExtensionAreaRoutes: readonly ExtensionAreaRoute[] = [
    ...extensionAreaRoutes,
    {
        path: '/-/manage',
        exact: true,
        render: lazyComponent(() => import('./RegistryExtensionManagePage'), 'RegistryExtensionManagePage'),
    },
    {
        path: '/-/releases/new',
        exact: true,
        render: lazyComponent(() => import('./RegistryExtensionNewReleasePage'), 'RegistryExtensionNewReleasePage'),
    },
]
