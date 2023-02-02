import { Redirect } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { namespaceAreaRoutes } from '../../namespaces/routes'
import { UserProfile } from '../profile/UserProfile'

import { UserAreaRoute } from './UserArea'

const UserSettingsArea = lazyComponent(() => import('../settings/UserSettingsArea'), 'UserSettingsArea')

const redirectToUserProfile: UserAreaRoute['render'] = props => <Redirect to={`${props.url}/profile`} />

export const userAreaRoutes: readonly UserAreaRoute[] = [
    {
        path: '/settings',
        render: props => (
            <UserSettingsArea
                {...props}
                routes={props.userSettingsAreaRoutes}
                sideBarItems={props.userSettingsSideBarItems}
            />
        ),
    },
    {
        path: '/profile',
        render: props => <UserProfile user={props.user} />,
    },
    ...namespaceAreaRoutes,

    // Redirect from /users/:username -> /users/:username/profile.
    {
        path: '/',
        exact: true,
        render: redirectToUserProfile,
    },
    // Redirect from previous /users/:username/account -> /users/:username/profile.
    {
        path: '/account',
        render: redirectToUserProfile,
    },
]
