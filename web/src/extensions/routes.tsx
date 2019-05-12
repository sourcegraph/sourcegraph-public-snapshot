import React from 'react'
import { asyncComponent } from '../util/asyncComponent'
import { ExtensionsAreaRoute } from './ExtensionsArea'

const ExtensionArea = asyncComponent(() => import('./extension/ExtensionArea'), 'ExtensionArea')

export const extensionsAreaRoutes: ReadonlyArray<ExtensionsAreaRoute> = [
    {
        path: '',
        exact: true,
        render: asyncComponent(() => import('./ExtensionsOverviewPage'), 'ExtensionsOverviewPage'),
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
