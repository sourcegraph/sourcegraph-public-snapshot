import * as React from 'react'
import { Link } from 'react-router-dom'
import { UserAvatar } from '../settings/user/UserAvatar'

export const NavLinks: React.SFC = () => (
    <div className="nav-links">
        {!window.context.onPrem && (
            <a href="https://about.sourcegraph.com" className="nav-links__link">
                About
            </a>
        )}
        {// if on-prem, never show a user avatar or sign-in button
        window.context.onPrem ? null : window.context.user ? (
            <Link className="nav-links__link" to="/settings">
                <UserAvatar size={64} />
            </Link>
        ) : (
            <Link className="nav-links__link btn btn-primary" to="/sign-in">
                Sign in
            </Link>
        )}
    </div>
)
