import React from 'react'

import * as GQL from '@sourcegraph/shared/src/schema'
import { Link } from '@sourcegraph/wildcard'

import { userURL } from '../../../user'

/**
 * Displays the account name as a link.
 */
export const AccountName: React.FunctionComponent<
    React.PropsWithChildren<{
        account: Pick<GQL.IUser, 'username' | 'displayName'> | null
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
