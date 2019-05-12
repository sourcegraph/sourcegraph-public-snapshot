import { ExtensionsAreaRoute } from '../../extensions/ExtensionsArea'
import { extensionsAreaRoutes } from '../../extensions/routes'
import { asyncComponent } from '../../util/asyncComponent'

export const enterpriseExtensionsAreaRoutes: ReadonlyArray<ExtensionsAreaRoute> = [
    extensionsAreaRoutes[0],
    {
        path: `/registry`,
        render: asyncComponent(() => import('./registry/RegistryArea'), 'RegistryArea'),
    },
    ...extensionsAreaRoutes.slice(1),
]
