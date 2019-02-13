import React from 'react'
import { eventLogger } from '../../tracking/eventLogger'
import { ExtensionAreaRoute } from './ExtensionArea'
const RegistryExtensionContributionsPage = React.lazy(async () => ({
    default: (await import('./RegistryExtensionContributionsPage')).RegistryExtensionContributionsPage,
}))
const RegistryExtensionManifestPage = React.lazy(async () => ({
    default: (await import('./RegistryExtensionManifestPage')).RegistryExtensionManifestPage,
}))
const RegistryExtensionOverviewPage = React.lazy(async () => ({
    default: (await import('./RegistryExtensionOverviewPage')).RegistryExtensionOverviewPage,
}))

export const extensionAreaRoutes: ReadonlyArray<ExtensionAreaRoute> = [
    {
        path: '',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <RegistryExtensionOverviewPage eventLogger={eventLogger} {...props} />,
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
