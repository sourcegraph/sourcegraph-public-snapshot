import React from 'react'

import { ExtensionsAreaRoute } from '../../extensions/registry/ExtensionsArea'
import { lazyComponent } from '../../util/lazyComponent'

const ExtensionArea = lazyComponent(() => import('./extension/ExtensionArea'), 'ExtensionArea')

export const enterpriseExtensionsAreaRoutes: readonly ExtensionsAreaRoute[] = [
    {
        path: '',
        exact: true,
        render: lazyComponent(() => import('./registry/ExtensionRegistry'), 'ExtensionRegistry'),
    },
    {
        path: '/registry',
        render: lazyComponent(() => import('./registry/RegistryArea'), 'RegistryArea'),
    },
    {
        path: '/:extensionID(.*)/-/',
        render: props => <ExtensionArea {...props} routes={props.extensionAreaRoutes} />,
    },
    {
        path: '/:extensionID(.*)',
        render: props => <ExtensionArea {...props} routes={props.extensionAreaRoutes} />,
    },
]
