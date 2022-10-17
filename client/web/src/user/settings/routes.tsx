import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { Text } from '@sourcegraph/wildcard'

import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'

import { showPasswordsPage, showAccountSecurityPage } from './cloud-ga'
import { UserSettingsAreaRoute } from './UserSettingsArea'

const SettingsArea = lazyComponent(() => import('../../settings/SettingsArea'), 'SettingsArea')

const UserSettingsSecurityPage = lazyComponent(
    () => import('./auth/UserSettingsSecurityPage'),
    'UserSettingsSecurityPage'
)

export const userSettingsAreaRoutes: readonly UserSettingsAreaRoute[] = [
    {
        path: '',
        exact: true,
        render: props => {
            if (props.isSourcegraphDotCom && props.authenticatedUser && props.user.id !== props.authenticatedUser.id) {
                return (
                    <SiteAdminAlert className="sidebar__alert" variant="danger">
                        Only the user may access their individual settings.
                    </SiteAdminAlert>
                )
            }
            return (
                <SettingsArea
                    {...props}
                    subject={props.user}
                    isLightTheme={props.isLightTheme}
                    extraHeader={
                        <>
                            {props.authenticatedUser && props.user.id !== props.authenticatedUser.id && (
                                <SiteAdminAlert className="sidebar__alert">
                                    Viewing settings for <strong>{props.user.username}</strong>
                                </SiteAdminAlert>
                            )}
                            <Text>User settings override global and organization settings.</Text>
                        </>
                    }
                />
            )
        },
    },
    {
        path: '/profile',
        exact: true,
        render: lazyComponent(() => import('./profile/UserSettingsProfilePage'), 'UserSettingsProfilePage'),
    },
    {
        path: '/password',
        exact: true,
        render: lazyComponent(() => import('./auth/UserSettingsPasswordPage'), 'UserSettingsPasswordPage'),
        condition: showPasswordsPage,
    },
    {
        path: '/emails',
        exact: true,
        render: lazyComponent(() => import('./emails/UserSettingsEmailsPage'), 'UserSettingsEmailsPage'),
    },
    {
        path: '/tokens',
        render: lazyComponent(() => import('./accessTokens/UserSettingsTokensArea'), 'UserSettingsTokensArea'),
        condition: () => window.context.accessTokensAllow !== 'none',
    },
    // future GA Cloud routes
    {
        path: '/security',
        exact: true,
        render: props => <UserSettingsSecurityPage {...props} context={window.context} />,
        condition: showAccountSecurityPage,
    },
    {
        path: '/product-research',
        exact: true,
        render: lazyComponent(() => import('./research/ProductResearch'), 'ProductResearchPage'),
        condition: () => window.context.productResearchPageEnabled,
    },
    {
        path: '/about-organizations',
        exact: true,
        render: lazyComponent(() => import('./aboutOrganization/AboutOrganizationPage'), 'AboutOrganizationPage'),
    },
]
