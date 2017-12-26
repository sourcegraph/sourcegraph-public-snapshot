import * as H from 'history'
import * as React from 'react'
import { VALID_PASSWORD_REGEXP } from '../settings/validation'

export const PasswordInput: React.SFC<React.InputHTMLAttributes<HTMLInputElement>> = props => (
    <input
        {...props}
        className={`form-control ${props.className || ''}`}
        type="password"
        placeholder="Password"
        pattern={VALID_PASSWORD_REGEXP.toString().slice(1, -1)}
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

/**
 * Returns the sanitized return-to URL, which is the location that a user
 * should be returned to after performing signin or signup to continue
 * to the page they intended to view as an authenticated user.
 *
 * 🚨 SECURITY: We must disallow open redirects (to arbitrary hosts).
 */
export function getReturnTo(location: H.Location): string | null {
    const searchParams = new URLSearchParams(location.search)
    const returnTo = searchParams.get('returnTo')
    if (returnTo) {
        const newURL = new URL(returnTo, window.location.href)
        return window.context.appURL + newURL.pathname + newURL.search + newURL.hash
    }
    return null
}
