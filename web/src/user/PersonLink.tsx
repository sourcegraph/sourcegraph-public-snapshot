import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'

/**
 * A person's name, with a link to their Sourcegraph user profile if an associated user is found.
 */
export const PersonLink: React.SFC<{
    displayName: string
    user?: GQL.IUser | null
    className?: string
    userClassName?: string
}> = ({ displayName, user, className = '', userClassName }) => (
    <>
        <span className={className}>{displayName}</span>
        {user && (
            <>
                {' '}
                &mdash;{' '}
                <Link
                    to={user.url}
                    className={`${className} ${userClassName || ''}`}
                    title={user.displayName || displayName}
                >
                    {user.username}
                </Link>
            </>
        )}
    </>
)
