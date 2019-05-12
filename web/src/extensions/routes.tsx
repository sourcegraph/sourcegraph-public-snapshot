import React from 'react'
import { asyncComponent } from '../util/asyncComponent'
import { ExtensionsAreaRoute } from './ExtensionsArea'

const ExtensionArea = asyncComponent(() => import('./extension/ExtensionArea'), 'ExtensionArea')
const ExtensionsOverviewPage = asyncComponent(() => import('./ExtensionsOverviewPage'), 'ExtensionsOverviewPage')

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
