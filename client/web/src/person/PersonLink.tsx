import * as React from 'react'

import classNames from 'classnames'

import { gql } from '@sourcegraph/http-client'
import { Tooltip, LinkOrSpan } from '@sourcegraph/wildcard'

import type { PersonLinkFields } from '../graphql-operations'

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
 * Formats a person name to display in the UI.
 * If the person has a user account, the user's display name is used if it exsits, otherwise the username is used.
 * If the person does not have a user account, the display name is used if it exists, otherwise the email is used.
 */
export const formatPersonName = ({ user, displayName, email }: PersonLinkFields): string =>
    user ? user.displayName || user.username : displayName || email

const formatTooltip = ({ user, email }: PersonLinkFields): string =>
    user ? `${user.username} ${email ? `<${email}>` : ''}` : email

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
    <Tooltip content={formatTooltip(person)}>
        <LinkOrSpan to={person.user?.url} className={classNames(className, person.user && userClassName)}>
            {formatPersonName(person)}
        </LinkOrSpan>
    </Tooltip>
)
