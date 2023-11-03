import { Navigate } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { namespaceAreaRoutes } from '../../namespaces/routes'

import type { UserAreaRoute } from './UserArea'

const UserSettingsArea = lazyComponent(() => import('../settings/UserSettingsArea'), 'UserSettingsArea')
const UserProfile = lazyComponent(() => import('../profile/UserProfile'), 'UserProfile')

export const userAreaRoutes: readonly UserAreaRoute[] = [
    {
        path: 'settings/*',
        render: props => (
            <UserSettingsArea
                {...props}
                routes={props.userSettingsAreaRoutes}
                sideBarItems={props.userSettingsSideBarItems}
                telemetryRecorder={props.platformContext.telemetryRecorder}
            />
        ),
    },
    {
        path: 'profile',
        render: props => <UserProfile user={props.user} />,
    },
    ...namespaceAreaRoutes,

    // Redirect from /users/:username -> /users/:username/profile.
    {
        path: '',
        render: () => <Navigate to="profile" replace={true} />,
    },
    // Redirect from previous /users/:username/account -> /users/:username/profile.
    {
        path: 'account',
        render: () => <Navigate to="../profile" replace={true} />,
    },
]
