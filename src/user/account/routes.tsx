import React from 'react'
import { UserAccountAreaRoute } from './UserAccountArea'
import { UserAccountCreateAccessTokenPage } from './UserAccountCreateAccessTokenPage'
import { UserAccountEmailsPage } from './UserAccountEmailsPage'
import { UserAccountPasswordPage } from './UserAccountPasswordPage'
import { UserAccountProfilePage } from './UserAccountProfilePage'
import { UserAccountTokensPage } from './UserAccountTokensPage'

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
