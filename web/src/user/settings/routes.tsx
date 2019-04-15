import React from 'react'
import { UserAccountAreaRoute } from './UserSettingsArea'
const UserAccountCreateAccessTokenPage = React.lazy(async () => ({
    default: (await import('./accessTokens/UserSettingsCreateAccessTokenPage')).UserAccountCreateAccessTokenPage,
}))
const UserAccountEmailsPage = React.lazy(async () => ({
    default: (await import('./emails/UserSettingsEmailsPage')).UserAccountEmailsPage,
}))
const UserAccountPasswordPage = React.lazy(async () => ({
    default: (await import('./auth/UserSettingsPasswordPage')).UserAccountPasswordPage,
}))
const UserAccountProfilePage = React.lazy(async () => ({
    default: (await import('./profile/UserSettingsProfilePage')).UserAccountProfilePage,
}))
const UserAccountTokensPage = React.lazy(async () => ({
    default: (await import('./accessTokens/UserSettingsTokensPage')).UserAccountTokensPage,
}))

export const userAccountAreaRoutes: ReadonlyArray<UserAccountAreaRoute> = [
    // Render empty page if no settings page selected
    {
        path: '/profile',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserAccountProfilePage {...props} />,
    },
    {
        path: '/password',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserAccountPasswordPage {...props} />,
    },
    {
        path: '/emails',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserAccountEmailsPage {...props} />,
    },
    {
        path: '/tokens',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserAccountTokensPage {...props} />,
        condition: () => window.context.accessTokensAllow !== 'none',
    },
    {
        path: '/tokens/new',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserAccountCreateAccessTokenPage {...props} />,
        condition: () => window.context.accessTokensAllow !== 'none',
    },
]
