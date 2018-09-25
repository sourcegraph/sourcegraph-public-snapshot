import { ExtensionAreaRoute } from '@sourcegraph/webapp/dist/extensions/extension/ExtensionArea'
import { extensionAreaRoutes } from '@sourcegraph/webapp/dist/extensions/extension/routes'
import React from 'react'
import { RegistryExtensionOverviewPage } from '../registry/RegistryExtensionOverviewPage'
import { RegistryExtensionManagePage } from './RegistryExtensionManagePage'
import { RegistryExtensionNewReleasePage } from './RegistryExtensionNewReleasePage'

export const enterpriseExtensionAreaRoutes: ReadonlyArray<ExtensionAreaRoute> = [
    {
        path: '',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <RegistryExtensionOverviewPage {...props} />,
    },
    ...extensionAreaRoutes,
    {
        path: `/-/manage`,
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <RegistryExtensionManagePage {...props} />,
    },
    {
        path: `/-/releases/new`,
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <RegistryExtensionNewReleasePage {...props} />,
    },
]
