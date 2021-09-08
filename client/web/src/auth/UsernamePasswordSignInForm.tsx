import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback, useState } from 'react'
import { Link } from 'react-router-dom'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { asError } from '@sourcegraph/shared/src/util/errors'

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
export const UsernamePasswordSignInForm: React.FunctionComponent<Props> = ({
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
                <div className="form-group d-flex flex-column align-content-start">
                    <label htmlFor="username-or-email" className="align-self-start">
                        Username or email
                    </label>
                    <input
                        id="username-or-email"
                        className="form-control"
                        type="text"
                        onChange={onUsernameOrEmailFieldChange}
                        required={true}
                        value={usernameOrEmail}
                        disabled={loading}
                        autoCapitalize="off"
                        autoFocus={true}
                        // There is no well supported way to declare username OR email here.
                        // Using username seems to be the best approach and should still support this behaviour.
                        // See: https://github.com/whatwg/html/issues/4445
                        autoComplete="username"
                    />
                </div>
                <div className="form-group d-flex flex-column align-content-start">
                    <div className="d-flex justify-content-between">
                        <label htmlFor="password">Password</label>
                        {context.resetPasswordEnabled && (
                            <small className="form-text text-muted">
                                <Link to="/password-reset">Forgot password?</Link>
                            </small>
                        )}
                    </div>
                    <PasswordInput
                        onChange={onPasswordFieldChange}
                        value={password}
                        required={true}
                        disabled={loading}
                        autoComplete="current-password"
                        placeholder=" "
                    />
                </div>
                <div
                    className={classNames('form-group', {
                        'mb-0': noThirdPartyProviders,
                    })}
                >
                    <button className="btn btn-primary btn-block" type="submit" disabled={loading}>
                        {loading ? <LoadingSpinner className="icon-inline" /> : 'Sign in'}
                    </button>
                </div>
            </Form>
        </>
    )
}
