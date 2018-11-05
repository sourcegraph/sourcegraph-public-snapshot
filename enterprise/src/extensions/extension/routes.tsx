import React from 'react'
import { ExtensionAreaRoute } from '../../../../packages/webapp/src/extensions/extension/ExtensionArea'
import { extensionAreaRoutes } from '../../../../packages/webapp/src/extensions/extension/routes'
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
