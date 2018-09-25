import { ExtensionsAreaRoute } from '@sourcegraph/webapp/dist/extensions/ExtensionsArea'
import { extensionsAreaRoutes } from '@sourcegraph/webapp/dist/extensions/routes'
import React from 'react'
import { RegistryArea } from './registry/RegistryArea'

export const enterpriseExtensionsAreaRoutes: ReadonlyArray<ExtensionsAreaRoute> = [
    extensionsAreaRoutes[0],
    {
        path: `/registry`,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <RegistryArea {...props} />,
    },
    ...extensionsAreaRoutes.slice(1),
]
