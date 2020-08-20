import React from 'react'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { lazyComponent } from '../../util/lazyComponent'
import { UserSettingsAreaRoute } from './UserSettingsArea'
import { codeHostExternalServices } from '../../components/externalServices/externalServices'

const SettingsArea = lazyComponent(() => import('../../settings/SettingsArea'), 'SettingsArea')
const ExternalServicesPage = lazyComponent(
    () => import('../../components/externalServices/ExternalServicesPage'),
    'ExternalServicesPage'
)
const AddExternalServicesPage = lazyComponent(
    () => import('../../components/externalServices/AddExternalServicesPage'),
    'AddExternalServicesPage'
)
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
        exact: true,
        render: lazyComponent(() => import('./accessTokens/UserSettingsTokensPage'), 'UserSettingsTokensPage'),
        condition: () => window.context.accessTokensAllow !== 'none',
    },
    {
        path: '/tokens/new',
        exact: true,
        render: lazyComponent(
            () => import('./accessTokens/UserSettingsCreateAccessTokenPage'),
            'UserSettingsCreateAccessTokenPage'
        ),
        condition: () => window.context.accessTokensAllow !== 'none',
    },
    {
        path: '/external-services',
        render: props => (
            <ExternalServicesPage
                {...props}
                userID={props.user.id}
                routingPrefix={props.user.url + '/settings'}
                afterDeleteRoute={props.user.url + '/settings/external-services'}
            />
        ),
        exact: true,
    },
    {
        path: '/external-services/new',
        render: props => (
            <AddExternalServicesPage
                {...props}
                routingPrefix={props.user.url + '/settings'}
                afterCreateRoute={props.user.url + '/settings/external-services'}
                userID={props.user.id}
                codeHostExternalServices={{
                    github: codeHostExternalServices.github,
                    gitlabcom: codeHostExternalServices.gitlabcom,
                    bitbucket: codeHostExternalServices.bitbucket,
                }}
                nonCodeHostExternalServices={{}}
            />
        ),
        exact: true,
    },
    {
        path: '/external-services/:id',
        render: props => (
            <ExternalServicePage {...props} afterUpdateRoute={props.user.url + '/settings/external-services'} />
        ),
        exact: true,
    },
]
