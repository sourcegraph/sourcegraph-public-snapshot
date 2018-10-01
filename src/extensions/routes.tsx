import React from 'react'
import { ExtensionArea } from './extension/ExtensionArea'
import { ExtensionsAreaRoute } from './ExtensionsArea'
import { ExtensionsOverviewPage } from './ExtensionsOverviewPage'

export const extensionsAreaRoutes: ReadonlyArray<ExtensionsAreaRoute> = [
    {
        path: '',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <ExtensionsOverviewPage {...props} />,
    },
    {
        path: `/:extensionID(.*)/-/`,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <ExtensionArea {...props} routes={props.extensionAreaRoutes} />,
    },
    {
        path: `/:extensionID(.*)`,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <ExtensionArea {...props} routes={props.extensionAreaRoutes} />,
    },
]
