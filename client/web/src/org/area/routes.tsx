import { Redirect } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { namespaceAreaRoutes } from '../../namespaces/routes'

import { OrgAreaRoute } from './OrgArea'

const OrgSettingsArea = lazyComponent(() => import('../settings/OrgSettingsArea'), 'OrgSettingsArea')
const OrgMembersArea = lazyComponent(() => import('../members/OrgMembersArea'), 'OrgMembersArea')
const OpenBetaGetStartedPage = lazyComponent(() => import('../openBeta/GettingStarted'), 'OpenBetaGetStartedPage')

const redirectToOrganizationProfile: OrgAreaRoute['render'] = props => (
    <Redirect to={`${props.match.url}/settings/profile`} />
)

export const orgAreaRoutes: readonly OrgAreaRoute[] = [
    {
        path: '/getstarted',
        render: props => <OpenBetaGetStartedPage {...props} />,
    },
    {
        path: '/settings/members',
        condition: context => context.newMembersInviteEnabled,
        render: props => <OrgMembersArea {...props} isLightTheme={props.isLightTheme} />,
    },
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
