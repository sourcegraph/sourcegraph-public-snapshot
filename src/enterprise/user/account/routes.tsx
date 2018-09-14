import React from 'react'
import { userAccountAreaRoutes } from '../../../user/account/routes'
import { UserAccountAreaRoute } from '../../../user/account/UserAccountArea'
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
