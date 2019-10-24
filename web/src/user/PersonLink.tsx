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
     * The Sourcegraph user, or if there is no user account associated with this person, then the
     * person's display name (as a string).
     */
    user: Pick<GQL.IUser, 'displayName' | 'username' | 'url'> | string

    className?: string

    /** A class name applied when there is an associated user account. */
    userClassName?: string
}> = ({ user, className = '', userClassName = '' }) =>
    typeof user === 'string' ? (
        <span className={className}>{user}</span>
    ) : (
        <Link to={user.url} className={`${className} ${userClassName}`} title={user.displayName || ''}>
            {user.username}
        </Link>
    )
