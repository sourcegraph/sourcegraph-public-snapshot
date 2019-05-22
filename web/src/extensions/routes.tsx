import React from 'react'
import { lazyComponent } from '../util/lazyComponent'
import { ExtensionsAreaRoute } from './ExtensionsArea'

const ExtensionArea = lazyComponent(() => import('./extension/ExtensionArea'), 'ExtensionArea')

export const extensionsAreaRoutes: ReadonlyArray<ExtensionsAreaRoute> = [
    {
        path: '',
        exact: true,
        render: lazyComponent(() => import('./ExtensionsOverviewPage'), 'ExtensionsOverviewPage'),
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
