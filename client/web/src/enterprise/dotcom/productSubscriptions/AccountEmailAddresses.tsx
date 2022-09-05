import React from 'react'

import classNames from 'classnames'

import * as GQL from '@sourcegraph/shared/src/schema'
import { Link } from '@sourcegraph/wildcard'

/**
 * Displays an inline list of email addresses for an account.
 */
export const AccountEmailAddresses: React.FunctionComponent<
    React.PropsWithChildren<{
        emails: Pick<GQL.IUserEmail, 'email' | 'verified'>[]
    }>
> = ({ emails }) => (
    <>
        {emails.map(({ email, verified }, index) => (
            <span
                key={index}
                className={classNames('text-nowrap d-inline-block mr-2', !verified && 'text-muted font-italic')}
            >
                <Link to={`mailto:${email}`}>{email}</Link> {verified ? '(verified)' : '(unverified)'}
            </span>
        ))}
    </>
)
