import * as React from 'react'

import classNames from 'classnames'
import * as H from 'history'

import { USERNAME_MAX_LENGTH, VALID_USERNAME_REGEXP } from '../user'

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
    inputRef?: React.Ref<HTMLInputElement>
}

export const PasswordInput: React.FunctionComponent<React.PropsWithChildren<InputProps>> = props => {
    const { inputRef, ...other } = props
    return (
        <input
            name="password"
            id="password"
            {...other}
            className={classNames('form-control', props.className)}
            placeholder={props.placeholder || 'Password'}
            type="password"
            required={true}
            ref={inputRef}
        />
    )
}

export const EmailInput: React.FunctionComponent<React.PropsWithChildren<InputProps>> = props => {
    const { inputRef, ...other } = props
    return (
        <input
            name="email"
            id="email"
            {...other}
            className={classNames('form-control', props.className)}
            type="email"
            placeholder={props.placeholder || 'Email'}
            spellCheck={false}
            autoComplete="email"
            ref={inputRef}
        />
    )
}

export const UsernameInput: React.FunctionComponent<React.PropsWithChildren<InputProps>> = props => {
    const { inputRef, ...other } = props
    return (
        <input
            name="username"
            id="username"
            {...other}
            className={classNames('form-control', props.className)}
            type="text"
            placeholder={props.placeholder || 'Username'}
            spellCheck={false}
            pattern={VALID_USERNAME_REGEXP}
            maxLength={USERNAME_MAX_LENGTH}
            autoCapitalize="off"
            autoComplete="username"
            ref={inputRef}
        />
    )
}

/**
 * Returns the sanitized return-to relative URL (including only the path, search, and fragment).
 * This is the location that a user should be returned to after performing signin or signup to continue
 * to the page they intended to view as an authenticated user.
 *
 * 🚨 SECURITY: We must disallow open redirects (to arbitrary hosts).
 */
export function getReturnTo(location: H.Location): string {
    const searchParameters = new URLSearchParams(location.search)
    const returnTo = searchParameters.get('returnTo') || '/search'
    const newURL = new URL(returnTo, window.location.href)

    newURL.searchParams.append('toast', 'integrations')
    return newURL.pathname + newURL.search + newURL.hash
}

export function maybeAddPostSignUpRedirect(url?: string): string {
    const enablePostSignupFlow = window.context?.experimentalFeatures?.enablePostSignupFlow
    const isDotCom = window.context?.sourcegraphDotComMode
    const shouldAddRedirect = isDotCom && enablePostSignupFlow

    if (url) {
        if (shouldAddRedirect) {
            // second param to protect against relative urls
            const urlObject = new URL(url, window.location.href)

            urlObject.searchParams.append('redirect', '/welcome')
            return urlObject.toString()
        }

        return url
    }

    return shouldAddRedirect ? '/welcome' : ''
}
