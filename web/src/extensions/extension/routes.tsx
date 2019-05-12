import React from 'react'
import { eventLogger } from '../../tracking/eventLogger'
import { lazyComponent } from '../../util/lazyComponent'
import { ExtensionAreaRoute } from './ExtensionArea'

const RegistryExtensionOverviewPage = lazyComponent(
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
        render: lazyComponent(() => import('./RegistryExtensionManifestPage'), 'RegistryExtensionManifestPage'),
    },
    {
        path: `/-/contributions`,
        exact: true,
        render: lazyComponent(
            () => import('./RegistryExtensionContributionsPage'),
            'RegistryExtensionContributionsPage'
        ),
    },
]
