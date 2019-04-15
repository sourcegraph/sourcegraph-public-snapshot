import React from 'react'
import { userSettingsAreaRoutes } from '../../../user/settings/routes'
import { UserSettingsAreaRoute } from '../../../user/settings/UserSettingsArea'
const UserSettingsExternalAccountsPage = React.lazy(async () => ({
    default: (await import('./UserSettingsExternalAccountsPage')).UserSettingsExternalAccountsPage,
}))

export const enterpriseUserSettingsAreaRoutes: ReadonlyArray<UserSettingsAreaRoute> = [
    ...userSettingsAreaRoutes,
    {
        path: '/external-accounts',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserSettingsExternalAccountsPage {...props} />,
    },
]
