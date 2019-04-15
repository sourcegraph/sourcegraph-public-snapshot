import React from 'react'
import { UserSettingsAreaRoute } from './UserSettingsArea'
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
    // Render empty page if no settings page selected
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
