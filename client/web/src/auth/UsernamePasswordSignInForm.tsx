import React, { useCallback, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError } from '@sourcegraph/common'
import { Label, Button, LoadingSpinner, Link, Text, Input } from '@sourcegraph/wildcard'

import { SourcegraphContext } from '../jscontext'
import { eventLogger } from '../tracking/eventLogger'

import { getReturnTo, PasswordInput } from './SignInSignUpCommon'

interface Props {
    location: H.Location
    history: H.History
    onAuthError: (error: Error | null) => void
    noThirdPartyProviders?: boolean
    context: Pick<
        SourcegraphContext,
        'allowSignup' | 'authProviders' | 'sourcegraphDotComMode' | 'xhrHeaders' | 'resetPasswordEnabled'
    >
}

/**
 * The form for signing in with a username and password.
 */
export const UsernamePasswordSignInForm: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    location,
    onAuthError,
    noThirdPartyProviders,
    context,
}) => {
    const [usernameOrEmail, setUsernameOrEmail] = useState('')
    const [password, setPassword] = useState('')
    const [loading, setLoading] = useState(false)

    const onUsernameOrEmailFieldChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setUsernameOrEmail(event.target.value)
    }, [])

    const onPasswordFieldChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setPassword(event.target.value)
    }, [])

    const handleSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>): void => {
            event.preventDefault()
            if (loading) {
                return
            }

            setLoading(true)
            eventLogger.log('InitiateSignIn')
            fetch('/-/sign-in', {
                credentials: 'same-origin',
                method: 'POST',
                headers: {
                    ...context.xhrHeaders,
                    Accept: 'application/json',
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    email: usernameOrEmail,
                    password,
                }),
            })
                .then(response => {
                    if (response.status === 200) {
                        if (new URLSearchParams(location.search).get('close') === 'true') {
                            window.close()
                        } else {
                            const returnTo = getReturnTo(location)
                            window.location.replace(returnTo)
                        }
                    } else if (response.status === 401) {
                        throw new Error('User or password was incorrect')
                    } else if (response.status === 422) {
                        throw new Error('The account has been locked out')
                    } else {
                        throw new Error('Unknown Error')
                    }
                })
                .catch(error => {
                    console.error('Auth error:', error)
                    setLoading(false)
                    onAuthError(asError(error))
                })
        },
        [usernameOrEmail, loading, location, password, onAuthError, context]
    )

    return (
        <>
            <Form onSubmit={handleSubmit}>
                <Input
                    id="username-or-email"
                    label={<Text alignment="left">Username or email</Text>}
                    onChange={onUsernameOrEmailFieldChange}
                    required={true}
                    value={usernameOrEmail}
                    disabled={loading}
                    autoCapitalize="off"
                    autoFocus={true}
                    className="form-group"
                    // There is no well supported way to declare username OR email here.
                    // Using username seems to be the best approach and should still support this behaviour.
                    // See: https://github.com/whatwg/html/issues/4445
                    autoComplete="username"
                />

                <div className="form-group d-flex flex-column align-content-start position-relative">
                    <Label htmlFor="password" className="align-self-start">
                        Password
                    </Label>
                    <PasswordInput
                        onChange={onPasswordFieldChange}
                        value={password}
                        required={true}
                        disabled={loading}
                        autoComplete="current-password"
                        placeholder=" "
                    />
                    {context.resetPasswordEnabled && (
                        <small className="form-text text-muted align-self-end position-absolute">
                            <Link to="/password-reset">Forgot password?</Link>
                        </small>
                    )}
                </div>

                <div
                    className={classNames('form-group', {
                        'mb-0': noThirdPartyProviders,
                    })}
                >
                    <Button className="btn-block" type="submit" disabled={loading} variant="primary">
                        {loading ? <LoadingSpinner /> : 'Sign in'}
                    </Button>
                </div>
            </Form>
        </>
    )
}
