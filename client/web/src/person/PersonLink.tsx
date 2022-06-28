import * as React from 'react'

import classNames from 'classnames'

import { gql } from '@sourcegraph/http-client'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'

import { PersonLinkFields } from '../graphql-operations'

export const personLinkFieldsFragment = gql`
    fragment PersonLinkFields on Person {
        email
        displayName
        user {
            username
            displayName
            url
        }
    }
`

/**
 * Formats a person name to: "username (Display Name)" or "Display Name"
 */
export const formatPersonName = ({ user, displayName }: PersonLinkFields): string =>
    user ? user.username : displayName

/**
 * A person's name, with a link to their Sourcegraph user profile if an associated user account is
 * found.
 */
export const PersonLink: React.FunctionComponent<
    React.PropsWithChildren<{
        /** The person to link to. */
        person: PersonLinkFields

        /** A class name that is always applied. */
        className?: string

        /** A class name applied when there is an associated user account. */
        userClassName?: string
    }>
> = ({ person, className = '', userClassName = '' }) => (
    <LinkOrSpan
        to={person.user?.url}
        className={classNames(className, person.user && userClassName)}
        data-tooltip={
            person.user && (person.user.displayName || person.displayName)
                ? `${person.user.displayName || person.displayName} <${person.email}>`
                : person.email
        }
    >
        {formatPersonName(person)}
    </LinkOrSpan>
)
