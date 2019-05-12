import React from 'react'
import { eventLogger } from '../../tracking/eventLogger'
import { asyncComponent } from '../../util/asyncComponent'
import { ExtensionAreaRoute } from './ExtensionArea'

const RegistryExtensionContributionsPage = asyncComponent(
    () => import('./RegistryExtensionContributionsPage'),
    'RegistryExtensionContributionsPage'
)
const RegistryExtensionManifestPage = asyncComponent(
    () => import('./RegistryExtensionManifestPage'),
    'RegistryExtensionManifestPage'
)
const RegistryExtensionOverviewPage = asyncComponent(
    () => import('./RegistryExtensionOverviewPage'),
    'RegistryExtensionOverviewPage'
)

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
