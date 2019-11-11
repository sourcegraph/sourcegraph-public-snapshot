import * as H from 'history'
import * as React from 'react'
import { USERNAME_MAX_LENGTH, VALID_USERNAME_REGEXP } from '../user'

export const PasswordInput: React.FunctionComponent<React.InputHTMLAttributes<HTMLInputElement> & {
    inputRef?: React.Ref<HTMLInputElement>
}> = props => {
    const { inputRef, ...other } = props
    return (
        <input
            name="password"
            {...other}
            className={`form-control ${props.className || ''}`}
            type="password"
            placeholder={props.placeholder === undefined ? 'Password' : props.placeholder}
            required={true}
            ref={inputRef}
        />
    )
}

export const EmailInput: React.FunctionComponent<React.InputHTMLAttributes<HTMLInputElement>> = props => (
    <input
        name="email"
        {...props}
        className={`form-control ${props.className || ''}`}
        type="email"
        placeholder="Email"
        spellCheck={false}
        autoComplete="email"
    />
)

export const UsernameInput: React.FunctionComponent<React.InputHTMLAttributes<HTMLInputElement>> = props => (
    <input
        name="username"
        {...props}
        className={`form-control ${props.className || ''}`}
        type="text"
        placeholder="Username"
        spellCheck={false}
        pattern={VALID_USERNAME_REGEXP}
        maxLength={USERNAME_MAX_LENGTH}
        autoCapitalize="off"
        autoComplete="username"
    />
)

/**
 * Returns the sanitized return-to relative URL (including only the path, search, and fragment).
 * This is the location that a user should be returned to after performing signin or signup to continue
 * to the page they intended to view as an authenticated user.
 *
 * 🚨 SECURITY: We must disallow open redirects (to arbitrary hosts).
 */
export function getReturnTo(location: H.Location): string {
    const searchParams = new URLSearchParams(location.search)
    const returnTo = searchParams.get('returnTo') || '/search'
    const newURL = new URL(returnTo, window.location.href)
    newURL.searchParams.append('toast', 'integrations')
    return newURL.pathname + newURL.search + newURL.hash
}
