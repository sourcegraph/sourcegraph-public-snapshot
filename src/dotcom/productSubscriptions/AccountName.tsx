import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { userURL } from '@sourcegraph/webapp/dist/user'
import React from 'react'
import { Link } from 'react-router-dom'

/**
 * Displays the account name as a link.
 */
export const AccountName: React.SFC<{
    account: Pick<GQL.IUser, 'username' | 'displayName'> | null
    link?: string
}> = ({ account, link }) =>
    account ? (
        <>
            <Link to={link || userURL(account.username)}>{account.username}</Link>{' '}
            {account.displayName && `(${account.displayName})`}
        </>
    ) : (
        <em>(Account deleted)</em>
    )
