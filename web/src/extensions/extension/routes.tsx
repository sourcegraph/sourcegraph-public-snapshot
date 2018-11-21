import React from 'react'
import { ExtensionAreaRoute } from './ExtensionArea'
import { RegistryExtensionContributionsPage } from './RegistryExtensionContributionsPage'
import { RegistryExtensionManifestPage } from './RegistryExtensionManifestPage'
import { RegistryExtensionOverviewPage } from './RegistryExtensionOverviewPage'

export const extensionAreaRoutes: ReadonlyArray<ExtensionAreaRoute> = [
    {
        path: '',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <RegistryExtensionOverviewPage {...props} />,
    },
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
