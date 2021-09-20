import { ExtensionAreaRoute } from '../../../extensions/extension/ExtensionArea'
import { lazyComponent } from '../../../util/lazyComponent'

export const enterpriseExtensionAreaRoutes: readonly ExtensionAreaRoute[] = [
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
