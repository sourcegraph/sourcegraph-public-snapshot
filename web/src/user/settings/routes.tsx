import React from 'react'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { UserSettingsAreaRoute } from './UserSettingsArea'
const SettingsArea = React.lazy(async () => ({
    default: (await import('../../settings/SettingsArea')).SettingsArea,
}))
const UserSettingsCreateAccessTokenPage = React.lazy(async () => ({
    default: (await import('./accessTokens/UserSettingsCreateAccessTokenPage')).UserSettingsCreateAccessTokenPage,
}))
const UserSettingsEmailsPage = React.lazy(async () => ({
    default: (await import('./emails/UserSettingsEmailsPage')).UserSettingsEmailsPage,
}))
const UserSettingsPasswordPage = React.lazy(async () => ({
    default: (await import('./auth/UserSettingsPasswordPage')).UserSettingsPasswordPage,
}))
const UserSettingsProfilePage = React.lazy(async () => ({
    default: (await import('./profile/UserSettingsProfilePage')).UserSettingsProfilePage,
}))
const UserSettingsTokensPage = React.lazy(async () => ({
    default: (await import('./accessTokens/UserSettingsTokensPage')).UserSettingsTokensPage,
}))

export const userSettingsAreaRoutes: ReadonlyArray<UserSettingsAreaRoute> = [
    {
        path: '',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
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
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserSettingsProfilePage {...props} />,
    },
    {
        path: '/password',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserSettingsPasswordPage {...props} />,
    },
    {
        path: '/emails',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserSettingsEmailsPage {...props} />,
    },
    {
        path: '/tokens',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserSettingsTokensPage {...props} />,
        condition: () => window.context.accessTokensAllow !== 'none',
    },
    {
        path: '/tokens/new',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserSettingsCreateAccessTokenPage {...props} />,
        condition: () => window.context.accessTokensAllow !== 'none',
    },
]
