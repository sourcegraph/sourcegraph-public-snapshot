import React from 'react'
import { ExtensionAreaRoute } from './ExtensionArea'
import { RegistryExtensionManifestPage } from './RegistryExtensionManifestPage'

export const extensionAreaRoutes: ReadonlyArray<ExtensionAreaRoute> = [
    {
        path: `/-/manifest`,
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <RegistryExtensionManifestPage {...props} />,
    },
]
