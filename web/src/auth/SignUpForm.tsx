import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import React, { useCallback, useState } from 'react'
import { from, Observable, Subscription } from 'rxjs'
import { asError } from '../../../shared/src/util/errors'
import { Form } from '../components/Form'
import { eventLogger } from '../tracking/eventLogger'
import { enterpriseTrial, signupTerms } from '../util/features'
import { EmailInput, PasswordInput, UsernameInput } from './SignInSignUpCommon'
import { ErrorAlert } from '../components/alerts'
import classNames from 'classnames'
import * as H from 'history'
import { OrDivider } from './OrDivider'
import GithubIcon from 'mdi-react/GithubIcon'
import { size } from 'lodash'
import { useEventObservable } from '../../../shared/src/util/useObservable'
import { map, switchMap } from 'rxjs/operators'

export interface SignUpArgs {
    email: string
    username: string
    password: string
    requestedTrial: boolean
}

interface SignUpFormProps {
    className?: string

    /** Called to perform the signup on the server. */
    doSignUp: (args: SignUpArgs) => Promise<void>

    buttonLabel?: string
    history: H.History
}
/**
 * TODO: Better naming
 */
interface SignUpFormValidator<T = string> {
    /**
     * Optional array of synchronous input validators.
     *
     * If there's no problem with the input, void return. Else,
     * return with the reason the input is invalid.
     */
    synchronousValidators?: ((value: T) => string | undefined)[]

    /**
     * Optional array of asynchronous input validators.
     *
     * If there's no problem with the input, void return. Else,
     * return with the reason the input is invalid.
     */
    asynchronousValidators?: ((value: T) => Promise<string | undefined>)[]
}

/** Lazily construct this in SignUpForm */
const signUpFormValidators: { [name: string]: SignUpFormValidator } = {
    email: {
        synchronousValidators: [],
        asynchronousValidators: [],
    },
    username: {
        synchronousValidators: [],
        asynchronousValidators: [],
    },
    password: {
        synchronousValidators: [],
    },
}

/**
 * TODO: RxJS integration w/ React component
 *
 * @param name
 * @param formValidator
 */
function validateFormInput<T>(value: T, name: string, formValidator: SignUpFormValidator<T>) {
    const { synchronousValidators, asynchronousValidators } = formValidator

    if (synchronousValidators) {
        // looping over validators here because we only need the first reason of invalidity
        for (const validator of synchronousValidators) {
            const reason = validator(value)
            if (reason) {
                // TODO
                break
            }
        }
    }

    if (asynchronousValidators) {
        // wish I could promise.race but for rejected promises only
        Promise.all(asynchronousValidators.map(validator => validator(value)))
            .then(reasons => {
                // just need the first reason
                for (const reason of reasons) {
                    if (reason) {
                        // TODO
                        break
                    }
                }
            })
            .catch(() => {
                // noop TODO
            })
    }
}

/**
 *
 */
export const SignUpForm: React.FunctionComponent<SignUpFormProps> = ({ doSignUp, history, buttonLabel, className }) => {
    const [email, setEmail] = useState('')
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const [loading, setLoading] = useState(false)
    const [requestedTrial, setRequestedTrial] = useState(false)
    const [error, setError] = useState<Error | null>(null)

    const onEmailFieldChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setEmail(event.target.value)
    }, [])

    const onUsernameFieldChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setUsername(event.target.value)
    }, [])

    const onPasswordFieldChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setPassword(event.target.value)
    }, [])

    const onRequestTrialFieldChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setRequestedTrial(event.target.checked)
    }, [])

    const handleSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>): void => {
            event.preventDefault()
            if (loading) {
                return
            }

            setLoading(true)
            doSignUp({
                email: email || '',
                username,
                password,
                requestedTrial,
            }).catch(error => {
                setError(asError(error))
                setLoading(false)
            })
            eventLogger.log('InitiateSignUp')
        },
        [doSignUp, email, username, password, loading, requestedTrial]
    )

    return (
        <Form
            className={classNames(
                'signin-signup-form',
                'signup-form',
                'test-signup-form',
                'border rounded p-4',
                className
            )}
            onSubmit={handleSubmit}
        >
            {error && <ErrorAlert className="mb-3" error={error} history={history} />}
            <div className="form-group d-flex flex-column align-content-start">
                <label className="align-self-start">Email</label>
                <EmailInput
                    className="signin-signup-form__input"
                    onChange={onEmailFieldChange}
                    required={true}
                    value={email}
                    disabled={loading}
                    autoFocus={true}
                    placeholder=" "
                />
            </div>
            <div className="form-group d-flex flex-column align-content-start">
                <label className="align-self-start">Username</label>
                <UsernameInput
                    className="signin-signup-form__input"
                    onChange={onUsernameFieldChange}
                    value={username}
                    required={true}
                    disabled={loading}
                    placeholder=" "
                />
            </div>
            <div className="form-group d-flex flex-column align-content-start">
                <label className="align-self-start">Password</label>
                <PasswordInput
                    className="signin-signup-form__input"
                    onChange={onPasswordFieldChange}
                    value={password}
                    required={true}
                    disabled={loading}
                    autoComplete="new-password"
                    placeholder=" "
                />
            </div>
            {enterpriseTrial && (
                <div className="form-group">
                    <div className="form-check">
                        <label className="form-check-label">
                            <input className="form-check-input" type="checkbox" onChange={onRequestTrialFieldChange} />
                            Try Sourcegraph Enterprise free for 30 days{' '}
                            {/* eslint-disable-next-line react/jsx-no-target-blank */}
                            <a target="_blank" rel="noopener" href="https://about.sourcegraph.com/pricing">
                                <HelpCircleOutlineIcon className="icon-inline" />
                            </a>
                        </label>
                    </div>
                </div>
            )}
            <div className="form-group mb-0">
                <button className="btn btn-primary btn-block" type="submit" disabled={loading}>
                    {loading ? <LoadingSpinner className="icon-inline" /> : buttonLabel || 'Sign up'}
                </button>
            </div>
            {window.context.sourcegraphDotComMode && (
                <>
                    {size(window.context.authProviders) > 0 && <OrDivider className="my-4" />}
                    {window.context.authProviders?.map((provider, index) => (
                        // Use index as key because display name may not be unique. This is OK
                        // here because this list will not be updated during this component's lifetime.
                        /* eslint-disable react/no-array-index-key */
                        <div className="mb-2" key={index}>
                            <a href={provider.authenticationURL} className="btn btn-secondary btn-block">
                                {provider.displayName === 'GitHub' && <GithubIcon className="icon-inline" />} Continue
                                with {provider.displayName}
                            </a>
                        </div>
                    ))}
                </>
            )}

            {signupTerms && (
                <p className="mt-3 mb-0">
                    <small className="form-text text-muted">
                        By signing up, you agree to our {/* eslint-disable-next-line react/jsx-no-target-blank */}
                        <a href="https://about.sourcegraph.com/terms" target="_blank" rel="noopener">
                            Terms of Service
                        </a>{' '}
                        and {/* eslint-disable-next-line react/jsx-no-target-blank */}
                        <a href="https://about.sourcegraph.com/privacy" target="_blank" rel="noopener">
                            Privacy Policy
                        </a>
                        .
                    </small>
                </p>
            )}
        </Form>
    )
}
