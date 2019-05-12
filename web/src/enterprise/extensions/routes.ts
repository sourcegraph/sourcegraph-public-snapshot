import { ExtensionsAreaRoute } from '../../extensions/ExtensionsArea'
import { extensionsAreaRoutes } from '../../extensions/routes'
import { asyncComponent } from '../../util/asyncComponent'

export const enterpriseExtensionsAreaRoutes: ReadonlyArray<ExtensionsAreaRoute> = [
    extensionsAreaRoutes[0],
    {
        path: `/registry`,
        render: asyncComponent(
            () => import('./registry/RegistryArea'),
            'RegistryArea',
            require.resolveWeak('./registry/RegistryArea')
        ),
    },
    ...extensionsAreaRoutes.slice(1),
]
