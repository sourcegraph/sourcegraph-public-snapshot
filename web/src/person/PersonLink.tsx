import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../shared/src/graphql/schema'

/**
 * A person's name, with a link to their Sourcegraph user profile if an associated user is found. It
 * is intended to be used to display the a value of the GraphQL type Person, which represents a
 * person that may or may not have an associated user account on Sourcegraph.
 */
export const PersonLink: React.FunctionComponent<{
    /**
     * The person to link to. If there is no user account associated with this person, then the
     * person's display name is shown.
     */
    person: Pick<GQL.IPerson, 'displayName'> & { user: Pick<GQL.IUser, 'username' | 'displayName' | 'url'> | null }

    className?: string

    /** A class name applied when there is an associated user account. */
    userClassName?: string
}> = ({ person, className = '', userClassName = '' }) =>
    person.user ? (
        <Link
            to={person.user.url}
            className={`${className} ${userClassName}`}
            title={person.user.displayName || person.displayName}
        >
            {person.user.username}
        </Link>
    ) : (
        <span className={className}>{person.displayName}</span>
    )
