import React from 'react'
import { ExtensionsAreaRoute } from './ExtensionsArea'
const ExtensionArea = React.lazy(async () => ({ default: (await import('./extension/ExtensionArea')).ExtensionArea }))
const ExtensionsOverviewPage = React.lazy(async () => ({
    default: (await import('./ExtensionsOverviewPage')).ExtensionsOverviewPage,
}))

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
