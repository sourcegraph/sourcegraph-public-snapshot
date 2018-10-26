import { userAccountAreaRoutes } from '@sourcegraph/webapp/dist/user/account/routes'
import { UserAccountAreaRoute } from '@sourcegraph/webapp/dist/user/account/UserAccountArea'
import React from 'react'
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
