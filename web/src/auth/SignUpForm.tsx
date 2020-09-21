import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import React, { useCallback, useMemo, useState } from 'react'
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
import { compact, head, size } from 'lodash'
import { USERNAME_MAX_LENGTH, VALID_USERNAME_REGEXP } from '../user'
import { Observable, of, timer, zip } from 'rxjs'
import { useEventObservable } from '../../../shared/src/util/useObservable'
import { catchError, debounce, map, switchMap, tap } from 'rxjs/operators'
import { typingDebounceTime } from '../search/input/QueryInput'
import CheckIcon from 'mdi-react/CheckIcon'
import { fromFetch } from 'rxjs/fetch'
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
interface FieldValidators {
    /**
     * Optional array of synchronous input validators.
     *
     * If there's no problem with the input, void return. Else,
     * return with the reason the input is invalid.
     */
    synchronousValidators?: ((value: string) => string | void)[]

    /**
     * Optional array of asynchronous input validators. These must return
     * observables created with `fromFetch` for easy cancellation in `switchMap`.
     *
     * If there's no problem with the input, void return. Else,
     * return with the reason the input is invalid.
     */
    asynchronousValidators?: ((value: string) => Observable<string | undefined>)[]
}

/** Lazily construct this in SignUpForm */
const signUpFieldValidators: { [name in 'email' | 'username' | 'password']: FieldValidators } = {
    email: {
        synchronousValidators: [checkEmailFormat, checkEmailPattern],
        asynchronousValidators: [isEmailUnique],
    },
    username: {
        synchronousValidators: [checkUsernameLength, checkUsernamePattern],
        asynchronousValidators: [isUsernameUnique],
    },
    password: {
        synchronousValidators: [checkPasswordLength],
    },
}

type ValidationResult = { kind: 'VALID' } | { kind: 'INVALID'; reason: string }
type ValidationPipeline = (events: Observable<React.ChangeEvent<HTMLInputElement>>) => Observable<ValidationResult>

/**
 * TODO: RxJS integration w/ React component. Create pipeline for
 * useEventObservable? Wrap in useMemo
 *
 * To be consumed by `useEventObservable`
 *
 * @param name Name of input field, used for descriptive error messages
 * @param onInputChange Function to execute side-effects given the latest input value and loading state.
 * Typically used to set state in a React component.
 * @param fieldValidators
 */
function createValidationPipeline(
    name: string,
    onInputChange: (inputState: { value: string; loading: boolean }) => void,
    fieldValidators: FieldValidators
): ValidationPipeline {
    const { synchronousValidators = [], asynchronousValidators = [] } = fieldValidators

    /**
     * Validation Pipeline takes an observable<string> and returns an observable
     * of Validation Result
     */
    return function validationPipeline(events): Observable<ValidationResult> {
        return events.pipe(
            map(event => event.target.value),
            map(value => {
                // When there's a sync validation error or no async validators,
                // don't render spinner (prevent jitter), don't make async
                // validation requests (unnecessary, waste of resources)

                const syncReason = head(compact(synchronousValidators.map(validator => validator(value))))
                return { value, syncReason, loading: !syncReason && asynchronousValidators.length > 0 }
            }),
            tap(({ value, loading }) => onInputChange({ value, loading })),
            debounce(({ loading }) => timer(loading ? typingDebounceTime : 0)),
            switchMap(({ value, syncReason, loading }) => {
                if (!loading) {
                    return of(
                        syncReason ? { kind: 'INVALID' as const, reason: syncReason } : { kind: 'VALID' as const }
                    )
                }

                return zip(...asynchronousValidators.map(validator => validator(value))).pipe(
                    map(values => head(compact(values))),
                    map(reason => (reason ? { kind: 'INVALID' as const, reason } : { kind: 'VALID' as const })),
                    tap(() => onInputChange({ value, loading: false }))
                )
            }),
            catchError(() => of({ kind: 'INVALID' as const, reason: `Unknown error validating ${name}` }))
        )
    }
}

/**
 *
 */
export const SignUpForm: React.FunctionComponent<SignUpFormProps> = ({ doSignUp, history, buttonLabel, className }) => {
    const [loading, setLoading] = useState(false)
    const [requestedTrial, setRequestedTrial] = useState(false)
    const [error, setError] = useState<Error | null>(
        new Error(
            'This is a long error used to demonstrate the wrapping effect of the error alert component inside of this container'
        )
    )

    const [emailState, setEmailState] = useState({ value: '', loading: false })
    const [nextEmailFieldChange, emailValidationResult] = useEventObservable<
        React.ChangeEvent<HTMLInputElement>,
        ValidationResult
    >(useMemo(() => createValidationPipeline('email', setEmailState, signUpFieldValidators.email), []))

    const [usernameState, setUsernameState] = useState({ value: '', loading: false })
    const [nextUsernameFieldChange, usernameValidationResult] = useEventObservable<
        React.ChangeEvent<HTMLInputElement>,
        ValidationResult
    >(useMemo(() => createValidationPipeline('username', setUsernameState, signUpFieldValidators.username), []))

    const [passwordState, setPasswordState] = useState({ value: '', loading: false })
    const [nextPasswordFieldChange, passwordValidationResult] = useEventObservable<
        React.ChangeEvent<HTMLInputElement>,
        ValidationResult
    >(useMemo(() => createValidationPipeline('password', setPasswordState, signUpFieldValidators.password), []))

    const canRegister =
        emailValidationResult?.kind === 'VALID' &&
        usernameValidationResult?.kind === 'VALID' &&
        passwordValidationResult?.kind === 'VALID'

    const disabled = loading || !canRegister
    // const disabled = loading

    const handleSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>): void => {
            event.preventDefault()
            if (disabled) {
                return
            }

            setLoading(true)
            doSignUp({
                email: emailState.value,
                username: usernameState.value,
                password: passwordState.value,
                requestedTrial,
            }).catch(error => {
                setError(asError(error))
                setLoading(false)
            })
            eventLogger.log('InitiateSignUp')
        },
        [doSignUp, disabled, emailState, usernameState, passwordState, requestedTrial]
    )

    const onRequestTrialFieldChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setRequestedTrial(event.target.checked)
    }, [])

    return (
        <>
            {error && <ErrorAlert className="mt-4 mb-0" error={error} history={history} />}
            <Form
                className={classNames(
                    'signin-signup-form',
                    'signup-form',
                    'test-signup-form',
                    'border rounded p-4',
                    'text-left',
                    window.context.sourcegraphDotComMode || error ? 'mt-3' : 'mt-4',
                    className
                )}
                onSubmit={handleSubmit}
                noValidate={true}
            >
                <div className="form-group d-flex flex-column align-content-start">
                    <label
                        className={classNames('align-self-start', {
                            'text-danger font-weight-bold':
                                emailValidationResult?.kind === 'INVALID' && !emailState.loading,
                        })}
                    >
                        Email
                    </label>
                    <div className="signin-signup-form__input-container">
                        <EmailInput
                            className={classNames('signin-signup-form__input', {
                                'border-danger': emailValidationResult?.kind === 'INVALID' && !emailState.loading,
                            })}
                            onChange={nextEmailFieldChange}
                            required={true}
                            value={emailState.value}
                            disabled={loading}
                            autoFocus={true}
                            placeholder=" "
                        />
                        {emailState.loading ? (
                            <LoadingSpinner className="signin-signup-form__icon" />
                        ) : (
                            emailValidationResult?.kind === 'VALID' && (
                                <CheckIcon className="signin-signup-form__icon text-success" size={20} />
                            )
                        )}
                    </div>
                    {!emailState.loading && emailValidationResult?.kind === 'INVALID' && (
                        <span className="text-danger signin-signup-form__subtext mt-2">
                            {emailValidationResult.reason}
                        </span>
                    )}
                </div>
                <div className="form-group d-flex flex-column align-content-start">
                    <label
                        className={classNames('align-self-start', {
                            'text-danger font-weight-bold':
                                usernameValidationResult?.kind === 'INVALID' && !usernameState.loading,
                        })}
                    >
                        Username
                    </label>
                    <div className="signin-signup-form__input-container">
                        <UsernameInput
                            className={classNames('signin-signup-form__input', {
                                'border-danger': usernameValidationResult?.kind === 'INVALID' && !usernameState.loading,
                            })}
                            onChange={nextUsernameFieldChange}
                            value={usernameState.value}
                            required={true}
                            disabled={loading}
                            placeholder=" "
                        />
                        {usernameState.loading ? (
                            <LoadingSpinner className="signin-signup-form__icon" />
                        ) : (
                            usernameValidationResult?.kind === 'VALID' && (
                                <CheckIcon className="signin-signup-form__icon text-success" size={20} />
                            )
                        )}
                    </div>
                    {!usernameState.loading && usernameValidationResult?.kind === 'INVALID' && (
                        <span className="text-danger signin-signup-form__subtext mt-2">
                            {usernameValidationResult.reason}
                        </span>
                    )}
                </div>
                <div className="form-group d-flex flex-column align-content-start">
                    <label
                        className={classNames('align-self-start', {
                            'text-danger font-weight-bold':
                                passwordValidationResult?.kind === 'INVALID' && !passwordState.loading,
                        })}
                    >
                        Password
                    </label>
                    <div className="signin-signup-form__input-container">
                        <PasswordInput
                            className={classNames('signin-signup-form__input', {
                                'border-danger': passwordValidationResult?.kind === 'INVALID' && !passwordState.loading,
                            })}
                            onChange={nextPasswordFieldChange}
                            value={passwordState.value}
                            required={true}
                            disabled={loading}
                            autoComplete="new-password"
                            placeholder=" "
                        />
                        {passwordState.loading ? (
                            <LoadingSpinner className="signin-signup-form__icon" />
                        ) : (
                            passwordValidationResult?.kind === 'VALID' && (
                                <CheckIcon className="signin-signup-form__icon text-success" size={20} />
                            )
                        )}
                    </div>
                    {!passwordState.loading && passwordValidationResult?.kind === 'INVALID' ? (
                        <span className="text-danger signin-signup-form__subtext mt-2">
                            {passwordValidationResult.reason}
                        </span>
                    ) : (
                        <span className="signin-signup-form__subtext mt-2">At least 12 characters</span>
                    )}
                </div>
                {enterpriseTrial && (
                    <div className="form-group">
                        <div className="form-check">
                            <label className="form-check-label">
                                <input
                                    className="form-check-input"
                                    type="checkbox"
                                    onChange={onRequestTrialFieldChange}
                                />
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
                    <button className="btn btn-primary btn-block" type="submit" disabled={disabled}>
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
                                    {provider.displayName === 'GitHub' && <GithubIcon className="icon-inline" />}{' '}
                                    Continue with {provider.displayName}
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
        </>
    )
}

// Synchronous Validators

function checkPasswordLength(password: string): string | void {
    if (password.length < 12) {
        return 'Password must contain at least 12 characters'
    }
}

function checkUsernameLength(username: string): string | void {
    if (username.length > USERNAME_MAX_LENGTH) {
        console.log('username is long')
        return `Username is longer than ${USERNAME_MAX_LENGTH} characters`
    }
}

function checkUsernamePattern(username: string): string | void {
    const valid = new RegExp(VALID_USERNAME_REGEXP).test(username)
    if (!valid) {
        return "Username doesn't match the requested format"
    }
}

/**
 * Simple email format validation to catch the most glaring errors
 * and display helpful error messages
 */
function checkEmailFormat(email: string): string | void {
    const parts = email.trim().split('@')
    if (parts.length < 2) {
        return "Please include an '@' in the email address"
    }
    if (parts.length > 2) {
        return "A part following '@' should not contain the symbol '@'"
    }
}

/**
 * Catch-all regex for errors not caught by `checkEmailFormat`.
 * From emailregex.com
 */
function checkEmailPattern(email: string): string | void {
    if (
        // eslint-disable-next-line no-useless-escape
        !/^(([^\s"(),.:;<>@[\\\]]+(\.[^\s"(),.:;<>@[\\\]]+)*)|(".+"))@((\[(?:\d{1,3}\.){3}\d{1,3}])|(([\dA-Za-z\-]+\.)+[A-Za-z]{2,}))$/.test(
            email
        )
    ) {
        return 'Please enter a valid email'
    }
}

// Asynchronous Validators

function isEmailUnique(email: string): Observable<string | undefined> {
    return fromFetch(`/-/check-email-taken/${email}`).pipe(
        switchMap(response => {
            switch (response.status) {
                case 200:
                    return of(`The email '${email}' is taken.`)
                case 404:
                    // Email is unique
                    return of(undefined)

                default:
                    return of('Unknown error validating username')
            }
        }),
        catchError(() => of('Unknown error validating email'))
    )
}

function isUsernameUnique(username: string): Observable<string | undefined> {
    return fromFetch(`/-/check-username-taken/${username}`).pipe(
        switchMap(response => {
            switch (response.status) {
                case 200:
                    return of(`The username '${username}' is taken.`)
                case 404:
                    // Username is unique
                    return of(undefined)

                default:
                    return of('Unknown error validating username')
            }
        }),
        catchError(() => of('Unknown error validating username'))
    )
}
