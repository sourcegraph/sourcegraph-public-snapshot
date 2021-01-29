import React from 'react'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { lazyComponent } from '../../util/lazyComponent'
import { UserSettingsAreaRoute, UserSettingsAreaRouteContext } from './UserSettingsArea'
import { Scalars } from '../../graphql-operations'
import { RouteComponentProps } from 'react-router'
import type { UserAddCodeHostsPageContainerProps } from './UserAddCodeHostsPageContainer'

const SettingsArea = lazyComponent(() => import('../../settings/SettingsArea'), 'SettingsArea')

const UserSettingsRepositoriesPage = lazyComponent(
    () => import('./repositories/UserSettingsRepositoriesPage'),
    'UserSettingsRepositoriesPage'
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

export const userSettingsAreaRoutes: readonly UserSettingsAreaRoute[] = [
    {
        path: '',
        exact: true,
        render: props => (
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
        ),
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
    {
        path: '/repositories',
        render: props => (
            <UserSettingsRepositoriesPage
                {...props}
                userID={props.user.id}
                routingPrefix={props.user.url + '/settings'}
            />
        ),
        exact: true,
        condition: props =>
            window.context.externalServicesUserModeEnabled ||
            (props.user.id === props.authenticatedUser.id &&
                props.authenticatedUser.tags.includes('AllowUserExternalServicePublic')) ||
            props.user.tags?.includes('AllowUserExternalServicePublic'),
    },
    {
        path: '/repositories/manage',
        render: props => (
            <UserSettingsManageRepositoriesPage
                {...props}
                userID={props.user.id}
                routingPrefix={props.user.url + '/settings'}
            />
        ),
        exact: true,
        condition: props =>
            window.context.externalServicesUserModeEnabled ||
            (props.user.id === props.authenticatedUser.id &&
                props.authenticatedUser.tags.includes('AllowUserExternalServicePublic')) ||
            props.user.tags?.includes('AllowUserExternalServicePublic'),
    },
    {
        path: '/code-hosts',
        render: props => (
            <UserAddCodeHostsPageContainer
                userID={props.user.id}
                routingPrefix={props.user.url + '/settings'}
                history={props.history}
                match={props.match}
                location={props.location}
            />
        ),
        exact: true,
        condition: props =>
            window.context.externalServicesUserModeEnabled ||
            (props.user.id === props.authenticatedUser.id &&
                props.authenticatedUser.tags.includes('AllowUserExternalServicePublic')) ||
            props.user.tags?.includes('AllowUserExternalServicePublic'),
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
        condition: props =>
            window.context.externalServicesUserModeEnabled ||
            (props.user.id === props.authenticatedUser.id &&
                props.authenticatedUser.tags.includes('AllowUserExternalServicePublic')) ||
            props.user.tags?.includes('AllowUserExternalServicePublic'),
    },
]
