import React from 'react'
import { Redirect } from 'react-router'

import { UserAreaRoute } from './UserArea'

const UserOverviewPage = React.lazy(async () => ({
    default: (await import('./UserOverviewPage')).UserOverviewPage,
}))

const UserSettingsArea = React.lazy(async () => ({
    default: (await import('../settings/UserSettingsArea')).UserSettingsArea,
}))

export const userAreaRoutes: ReadonlyArray<UserAreaRoute> = [
    {
        path: '',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserOverviewPage {...props} />,
    },
    {
        path: '/settings',
        // tslint:disable-next-line:jsx-no-lambda
        render: props => (
            <UserSettingsArea
                {...props}
                routes={props.userSettingsAreaRoutes}
                sideBarItems={props.userSettingsSideBarItems}
                isLightTheme={props.isLightTheme}
            />
        ),
    },

    // Redirect from previous /users/:username/account -> /users/:username/settings/profile.
    {
        path: '/account',
        render: props => <Redirect to={`${props.url}/settings/profile`} />,
    },
]
