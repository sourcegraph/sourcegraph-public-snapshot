import { lazyComponent } from '../../util/lazyComponent'
import { ExtensionAreaRoute } from './ExtensionArea'

export const extensionAreaRoutes: readonly ExtensionAreaRoute[] = [
    {
        path: '',
        exact: true,
        render: lazyComponent(() => import('./RegistryExtensionOverviewPage'), 'RegistryExtensionOverviewPage'),
    },
    {
        path: '/-/manifest',
        exact: true,
        render: lazyComponent(() => import('./RegistryExtensionManifestPage'), 'RegistryExtensionManifestPage'),
    },
    {
        path: '/-/contributions',
        exact: true,
        render: lazyComponent(
            () => import('./RegistryExtensionContributionsPage'),
            'RegistryExtensionContributionsPage'
        ),
    },
]
