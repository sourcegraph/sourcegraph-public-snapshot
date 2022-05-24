import { RouteComponentProps } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { Scalars } from '../../graphql-operations'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'

import { showPasswordsPage, showAccountSecurityPage, userExternalServicesEnabled } from './cloud-ga'
import type { UserAddCodeHostsPageContainerProps } from './UserAddCodeHostsPageContainer'
import { UserSettingsAreaRoute, UserSettingsAreaRouteContext } from './UserSettingsArea'

const SettingsArea = lazyComponent(() => import('../../settings/SettingsArea'), 'SettingsArea')

const SettingsRepositoriesPage = lazyComponent(
    () => import('./repositories/SettingsRepositoriesPage'),
    'SettingsRepositoriesPage'
)
const UserSettingsManageRepositoriesPage = lazyComponent(
    () => import('./repositories/UserSettingsManageRepositoriesPage'),
    'UserSettingsManageRepositoriesPage'
)

const UserAddCodeHostsPageContainer = lazyComponent<
    UserAddCodeHostsPageContainerProps,
    'UserAddCodeHostsPageContainer'
>(() => import('./UserAddCodeHostsPageContainer'), 'UserAddCodeHostsPageContainer')

const ExternalServicePage = lazyComponent(
    () => import('../../components/externalServices/ExternalServicePage'),
    'ExternalServicePage'
)

const UserSettingsSecurityPage = lazyComponent(
    () => import('./auth/UserSettingsSecurityPage'),
    'UserSettingsSecurityPage'
)

// const UserSettingsPrivacyPage = lazyComponent(
//     () => import('./privacy/UserSettingsPrivacyPage'),
//     'UserSettingsPrivacyPage'
// )

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
                            <p>User settings override global and organization settings.</p>
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
        path: '/privacy',
        exact: true,
        render: lazyComponent(() => import('./privacy/UserSettingsPrivacyPage'), 'UserSettingsPrivacyPage'),
    },
    {
        path: '/repositories',
        render: props => (
            <SettingsRepositoriesPage
                {...props}
                owner={{ id: props.user.id, type: 'user', tags: props.authenticatedUser.tags }}
                routingPrefix={props.user.url + '/settings'}
            />
        ),
        exact: true,
        condition: userExternalServicesEnabled,
    },
    {
        path: '/repositories/manage',
        render: props => (
            <UserSettingsManageRepositoriesPage
                {...props}
                owner={{ id: props.authenticatedUser.id, tags: props.authenticatedUser.tags, type: 'user' }}
                routingPrefix={props.user.url + '/settings'}
            />
        ),
        exact: true,
        condition: userExternalServicesEnabled,
    },
    {
        path: '/code-hosts',
        render: props => (
            <UserAddCodeHostsPageContainer
                owner={{ id: props.authenticatedUser.id, tags: props.authenticatedUser.tags, type: 'user' }}
                context={window.context}
                routingPrefix={props.user.url + '/settings'}
                onUserExternalServicesOrRepositoriesUpdate={props.onUserExternalServicesOrRepositoriesUpdate}
                telemetryService={props.telemetryService}
            />
        ),
        exact: true,
        condition: userExternalServicesEnabled,
    },
    {
        path: '/external-services/:id',
        render: ({ match, ...props }: RouteComponentProps<{ id: Scalars['ID'] }> & UserSettingsAreaRouteContext) => (
            <ExternalServicePage
                {...props}
                externalServiceID={match.params.id}
                afterUpdateRoute={props.user.url + '/settings/external-services'}
            />
        ),
        exact: true,
        condition: userExternalServicesEnabled,
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
