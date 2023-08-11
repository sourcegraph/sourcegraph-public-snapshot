import { Navigate } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { namespaceAreaRoutes } from '../../namespaces/routes'

import type { OrgAreaRoute } from './OrgArea'

const OrgSettingsArea = lazyComponent(() => import('../settings/OrgSettingsArea'), 'OrgSettingsArea')

export const orgAreaRoutes: readonly OrgAreaRoute[] = [
    {
        path: 'getstarted',
        render: props => <Navigate to={`/organizations/${props.org.name}/settings/members`} replace={true} />,
    },
    {
        path: 'settings/*',
        render: props => (
            <OrgSettingsArea
                {...props}
                routes={props.orgSettingsAreaRoutes}
                sideBarItems={props.orgSettingsSideBarItems}
            />
        ),
    },
    ...namespaceAreaRoutes,

    // Redirect from /organizations/:orgname -> /organizations/:orgname/settings/profile.
    {
        path: '',
        render: () => <Navigate to="./settings/profile" />,
    },
    // Redirect from previous /organizations/:orgname/account -> /organizations/:orgname/settings/profile.
    {
        path: 'account',
        render: () => <Navigate to="../settings/profile" />,
    },
]
