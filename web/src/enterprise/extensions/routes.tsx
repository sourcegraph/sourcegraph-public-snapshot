import React from 'react'
import { ExtensionsAreaRoute } from '../../extensions/ExtensionsArea'
import { extensionsAreaRoutes } from '../../extensions/routes'
import { asyncComponent } from '../../util/asyncComponent'

const RegistryArea = asyncComponent(() => import('./registry/RegistryArea'), 'RegistryArea')

export const enterpriseExtensionsAreaRoutes: ReadonlyArray<ExtensionsAreaRoute> = [
    extensionsAreaRoutes[0],
    {
        path: `/registry`,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <RegistryArea {...props} />,
    },
    ...extensionsAreaRoutes.slice(1),
]
