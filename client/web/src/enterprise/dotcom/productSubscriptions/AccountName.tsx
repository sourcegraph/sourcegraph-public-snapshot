import React from 'react'

import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { RouterLink } from '@sourcegraph/wildcard'

import { userURL } from '../../../user'

/**
 * Displays the account name as a link.
 */
export const AccountName: React.FunctionComponent<{
    account: Pick<GQL.IUser, 'username' | 'displayName'> | null
    link?: string
}> = ({ account, link }) =>
    account ? (
        <>
            <RouterLink to={link || userURL(account.username)}>{account.username}</RouterLink>{' '}
            {account.displayName && `(${account.displayName})`}
        </>
    ) : (
        <em>(Account deleted)</em>
    )
