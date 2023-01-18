import { Redirect } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { namespaceAreaRoutes } from '../../namespaces/routes'

import { OrgAreaRoute } from './OrgArea'

const OrgSettingsArea = lazyComponent(() => import('../settings/OrgSettingsArea'), 'OrgSettingsArea')

const redirectToOrganizationProfile: OrgAreaRoute['render'] = props => (
    <Redirect to={`${props.match.url}/settings/profile`} />
)

export const orgAreaRoutes: readonly OrgAreaRoute[] = [
    {
        path: '/getstarted',
        render: props => <Redirect to={`/organizations/${props.org.name}/settings/members`} />,
    },
    {
        path: '/settings',
        render: props => (
            <OrgSettingsArea
                {...props}
                routes={props.orgSettingsAreaRoutes}
                sideBarItems={props.orgSettingsSideBarItems}
                isLightTheme={props.isLightTheme}
            />
        ),
    },
    ...namespaceAreaRoutes,

    // Redirect from /organizations/:orgname -> /organizations/:orgname/settings/profile.
    {
        path: '/',
        exact: true,
        render: redirectToOrganizationProfile,
    },
    // Redirect from previous /organizations/:orgname/account -> /organizations/:orgname/settings/profile.
    {
        path: '/account',
        render: redirectToOrganizationProfile,
    },
]
