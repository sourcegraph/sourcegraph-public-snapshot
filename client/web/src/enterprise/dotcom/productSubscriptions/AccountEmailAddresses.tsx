import React from 'react'

import classNames from 'classnames'

import { Link } from '@sourcegraph/wildcard'

import type { DotComProductSubscriptionEmailFields } from '../../../graphql-operations'

/**
 * Displays an inline list of email addresses for an account.
 */
export const AccountEmailAddresses: React.FunctionComponent<
    React.PropsWithChildren<{
        emails: DotComProductSubscriptionEmailFields['emails']
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
