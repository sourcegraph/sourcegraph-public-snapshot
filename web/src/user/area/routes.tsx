import React from 'react'
import { Redirect } from 'react-router'

import { UserSavedSearchesCreateForm } from '../saved-searches/UserSavedSearchesCreateForm'
import { UserSavedSearchesUpdateForm } from '../saved-searches/UserSavedSearchesUpdateForm'
import { UserSavedSearchListPage } from '../saved-searches/UserSavedSearchListPage'
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
    {
        path: '/searches',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserSavedSearchListPage {...props} />,
    },
    {
        path: '/searches/add',
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserSavedSearchesCreateForm {...props} />,
    },
    {
        path: '/searches/:id',
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserSavedSearchesUpdateForm {...props} />,
    },

    // Redirect from previous /users/:username/account -> /users/:username/settings/profile.
    {
        path: '/account',
        render: props => <Redirect to={`${props.url}/settings/profile`} />,
    },
]
