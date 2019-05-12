import React from 'react'
import { Redirect } from 'react-router'
import { asyncComponent } from '../../util/asyncComponent'
import { UserAreaRoute } from './UserArea'

const UserOverviewPage = asyncComponent(() => import('./UserOverviewPage'), 'UserOverviewPage')
const UserSettingsArea = asyncComponent(() => import('../settings/UserSettingsArea'), 'UserSettingsArea')

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
