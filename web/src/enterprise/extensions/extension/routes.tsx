import { ExtensionAreaRoute } from '../../../extensions/extension/ExtensionArea'
import { extensionAreaRoutes } from '../../../extensions/extension/routes'
import { asyncComponent } from '../../../util/asyncComponent'

export const enterpriseExtensionAreaRoutes: ReadonlyArray<ExtensionAreaRoute> = [
    ...extensionAreaRoutes,
    {
        path: `/-/manage`,
        exact: true,
        render: asyncComponent(() => import('./RegistryExtensionManagePage'), 'RegistryExtensionManagePage'),
    },
    {
        path: `/-/releases/new`,
        exact: true,
        render: asyncComponent(() => import('./RegistryExtensionNewReleasePage'), 'RegistryExtensionNewReleasePage'),
    },
]
