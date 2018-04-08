import * as React from 'react'
import { Link } from 'react-router-dom'

/**
 * A link to a user, or an unlinked person name.
 */
export const PersonLink: React.SFC<{
    displayName: string
    user?: GQL.IUser | null
    className?: string
    userClassName?: string
}> = ({ displayName, user, className = '', userClassName }) =>
    user ? (
        <Link to={user.url} className={`${className} ${userClassName || ''}`} title={user.displayName || displayName}>
            {user.username}
        </Link>
    ) : (
        <span className={className}>{displayName}</span>
    )
