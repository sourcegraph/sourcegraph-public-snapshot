import * as React from 'react'

import type * as H from 'history'

import { Input, type InputProps } from '@sourcegraph/wildcard'

import { PageRoutes } from '../routes.constants'
import { USERNAME_MAX_LENGTH, VALID_USERNAME_REGEXP } from '../user'

interface CommonInputProps extends InputProps, React.InputHTMLAttributes<HTMLInputElement> {
    inputRef?: React.Ref<HTMLInputElement>
}

export const PasswordInput: React.FunctionComponent<React.PropsWithChildren<CommonInputProps>> = props => {
    const { inputRef, ...other } = props
    return (
        <Input
            name="password"
            id="password"
            {...other}
            className={props.className}
            placeholder={props.placeholder || 'Password'}
            type="password"
            required={true}
            ref={inputRef}
        />
    )
}

export const EmailInput: React.FunctionComponent<React.PropsWithChildren<CommonInputProps>> = props => {
    const { inputRef, ...other } = props
    return (
        <Input
            name="email"
            id="email"
            {...other}
            className={props.className}
            type="email"
            placeholder={props.placeholder || 'Email'}
            spellCheck={false}
            autoComplete="email"
            ref={inputRef}
        />
    )
}

export const UsernameInput: React.FunctionComponent<React.PropsWithChildren<CommonInputProps>> = props => {
    const { inputRef, ...other } = props
    return (
        <Input
            name="username"
            id="username"
            {...other}
            className={props.className}
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
export function getReturnTo(location: H.Location, defaultReturnTo: string = PageRoutes.Search): string {
    const searchParameters = new URLSearchParams(location.search)
    const returnTo = searchParameters.get('returnTo') || defaultReturnTo

    // 🚨 SECURITY: check newURL scheme http or https or relative
    if (returnTo.startsWith('http:') || returnTo.startsWith('https:') || returnTo.startsWith('/')) {
        const newURL = new URL(returnTo, window.location.href)
        return newURL.pathname + newURL.search + newURL.hash
    }
    return defaultReturnTo
}
