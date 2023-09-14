import React from 'react'

import { Link } from '@sourcegraph/wildcard'

import type { ProductLicenseSubscriptionAccount } from '../../../graphql-operations'
import { userURL } from '../../../user'

/**
 * Displays the account name as a link.
 */
export const AccountName: React.FunctionComponent<
    React.PropsWithChildren<{
        account: Pick<ProductLicenseSubscriptionAccount, 'username' | 'displayName'> | null
        link?: string
    }>
> = ({ account, link }) =>
    account ? (
        <>
            <Link to={link || userURL(account.username)}>{account.username}</Link>{' '}
            {account.displayName && `(${account.displayName})`}
        </>
    ) : (
        <em>(Account deleted)</em>
    )
