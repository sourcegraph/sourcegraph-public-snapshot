import React from 'react'
import { userAccountAreaRoutes } from '../../../../web/src/user/account/routes'
import { UserAccountAreaRoute } from '../../../../web/src/user/account/UserAccountArea'
import { UserAccountExternalAccountsPage } from './UserAccountExternalAccountsPage'

export const enterpriseUserAccountAreaRoutes: ReadonlyArray<UserAccountAreaRoute> = [
    ...userAccountAreaRoutes,
    {
        path: '/external-accounts',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserAccountExternalAccountsPage {...props} />,
    },
]
