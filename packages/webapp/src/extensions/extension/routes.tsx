import React from 'react'
import { ExtensionAreaRoute } from './ExtensionArea'
import { RegistryExtensionContributionsPage } from './RegistryExtensionContributionsPage'
import { RegistryExtensionManifestPage } from './RegistryExtensionManifestPage'

export const extensionAreaRoutes: ReadonlyArray<ExtensionAreaRoute> = [
    {
        path: `/-/manifest`,
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <RegistryExtensionManifestPage {...props} />,
    },
    {
        path: `/-/contributions`,
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <RegistryExtensionContributionsPage {...props} />,
    },
]
