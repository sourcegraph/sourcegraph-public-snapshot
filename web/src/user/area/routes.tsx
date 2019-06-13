import React from 'react'
import { Redirect } from 'react-router'
import { lazyComponent } from '../../util/lazyComponent'
import { UserAreaRoute } from './UserArea'

const UserSettingsArea = lazyComponent(() => import('../settings/UserSettingsArea'), 'UserSettingsArea')

export const userAreaRoutes: ReadonlyArray<UserAreaRoute> = [
    {
        path: '',
        exact: true,
        render: lazyComponent(() => import('./UserOverviewPage'), 'UserOverviewPage'),
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
        render: lazyComponent(() => import('../saved-searches/UserSavedSearchListPage'), 'UserSavedSearchListPage'),
    },
    {
        path: '/searches/add',
        render: lazyComponent(
            () => import('../saved-searches/UserSavedSearchesCreateForm'),
            'UserSavedSearchesCreateForm'
        ),
    },
    {
        path: '/searches/:id',
        render: lazyComponent(
            () => import('../saved-searches/UserSavedSearchesUpdateForm'),
            'UserSavedSearchesUpdateForm'
        ),
    },
    {
        path: '/namespace',
        render: lazyComponent(() => import('../../namespaces/NamespaceArea'), 'NamespaceArea'),
    },

    // Redirect from previous /users/:username/account -> /users/:username/settings/profile.
    {
        path: '/account',
        render: props => <Redirect to={`${props.url}/settings/profile`} />,
    },
]
