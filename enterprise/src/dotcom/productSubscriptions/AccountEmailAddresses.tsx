import React from 'react'
import * as GQL from '../../../../packages/webapp/src/backend/graphqlschema'

/**
 * Displays an inline list of email addresses for an account.
 */
export const AccountEmailAddresses: React.SFC<{ emails: Pick<GQL.IUserEmail, 'email' | 'verified'>[] }> = ({
    emails,
}) => (
    <>
        {emails.map(({ email, verified }, i) => (
            <span key={i} className={`text-nowrap d-inline-block mr-2 ${verified ? '' : 'text-muted font-italic'}`}>
                <a href={`mailto:${email}`}>{email}</a> {verified ? '(verified)' : '(unverified)'}
            </span>
        ))}
    </>
)
