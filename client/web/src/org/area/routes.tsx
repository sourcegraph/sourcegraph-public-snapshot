import React from 'react'
import { Redirect } from 'react-router'

import { namespaceAreaRoutes } from '../../namespaces/routes'
import { lazyComponent } from '../../util/lazyComponent'

import { OrgAreaRoute } from './OrgArea'

const OrgSettingsArea = lazyComponent(() => import('../settings/OrgSettingsArea'), 'OrgSettingsArea')

const redirectToOrganizationProfile: OrgAreaRoute['render'] = props => (
    <Redirect to={`${props.match.url}/settings/profile`} />
)

export const orgAreaRoutes: readonly OrgAreaRoute[] = [
    {
        path: '/settings',
        render: props => <OrgSettingsArea {...props} isLightTheme={props.isLightTheme} />,
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
