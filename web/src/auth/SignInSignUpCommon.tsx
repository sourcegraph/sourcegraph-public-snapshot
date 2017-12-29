import * as H from 'history'
import * as React from 'react'
import { VALID_USERNAME_REGEXP } from '../settings/validation'

export const PasswordInput: React.SFC<React.InputHTMLAttributes<HTMLInputElement>> = props => (
    <input
        {...props}
        className={`form-control ${props.className || ''}`}
        type="password"
        placeholder="Password"
        minLength={window.context.debug ? 1 : 6}
    />
)

export const EmailInput: React.SFC<React.InputHTMLAttributes<HTMLInputElement>> = props => (
    <input
        {...props}
        className={`form-control ${props.className || ''}`}
        type="email"
        placeholder="Email"
        spellCheck={false}
    />
)

export const UsernameInput: React.SFC<React.InputHTMLAttributes<HTMLInputElement>> = props => (
    <input
        {...props}
        className={`form-control ${props.className || ''}`}
        type="text"
        placeholder="Username"
        spellCheck={false}
        pattern={VALID_USERNAME_REGEXP.toString().slice(1, -1)}
    />
)

/**
 * Returns the sanitized return-to relative URL (including only the path, search, and fragment).
 * This is the location that a user should be returned to after performing signin or signup to continue
 * to the page they intended to view as an authenticated user.
 *
 * 🚨 SECURITY: We must disallow open redirects (to arbitrary hosts).
 */
export function getReturnTo(location: H.Location): string | null {
    const searchParams = new URLSearchParams(location.search)
    const returnTo = searchParams.get('returnTo')
    if (returnTo) {
        const newURL = new URL(returnTo, window.location.href)
        return newURL.pathname + newURL.search + newURL.hash
    }
    return null
}
