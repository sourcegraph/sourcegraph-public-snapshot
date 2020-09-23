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
import { merge, Observable, of, partition, timer, zip } from 'rxjs'
import { useEventObservable } from '../../../shared/src/util/useObservable'
import { catchError, debounce, filter, map, switchMap, takeUntil, tap } from 'rxjs/operators'
import { typingDebounceTime } from '../search/input/QueryInput'
import CheckIcon from 'mdi-react/CheckIcon'
import { fromFetch } from 'rxjs/fetch'
import { isDefined } from '../../../shared/src/util/types'
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
 * Configuration used to create validation pipelines for `useEventObservable`
 *
 * TODO: Consider changing "OK" return type to false if this is ever used elsewehere.
 */
interface FieldValidators {
    /**
     * Optional array of synchronous input validators.
     *
     * If there's no problem with the input, return undefined. Else,
     * return with the reason the input is invalid.
     */
    synchronousValidators?: ((value: string) => string | undefined)[]

    /**
     * Optional array of asynchronous input validators. These must return
     * observables created with `fromFetch` for easy cancellation in `switchMap`.
     *
     * If there's no problem with the input, emit undefined. Else,
     * return with the reason the input is invalid.
     */
    asynchronousValidators?: ((value: string) => Observable<string | undefined>)[]
}

type ValidationEvent =
    | { kind: 'BLUR'; reason: string }
    | { kind: 'CHANGE_ASYNC'; value: string }
    | { kind: 'CHANGE_SYNC'; reason: string; value: string }
type ValidationResult = { kind: 'VALID' } | { kind: 'INVALID'; reason: string }
type ValidationPipeline = (
    events: Observable<React.ChangeEvent<HTMLInputElement> | React.FocusEvent<HTMLInputElement>>
) => Observable<ValidationResult>
interface InputState {
    value: string
    loading: boolean
}

/**
 * Returns an observable pipeline to be consumed by `useEventObservable`.
 * Helps with management of sync + async validation.
 * The returned pipeline takes both input change and focus events.
 *
 * Intended behavior:
 * - Asynchronous validation occurs on change (debounced)
 * - Synchronous validation occurs on blur (+ cancels any pending async validation requests)
 *
 * @param name Name of input field, used for descriptive error messages.
 * @param onInputChange Higher order function to execute side-effects given the latest input value and loading state.
 * Typically used to set state in a React component.
 * The function provided to `onInputChange` should be called with the previous input value and loading state
 * @param fieldValidators Config object that declares sync + async validators
 */
function createValidationPipeline(
    name: string,
    onInputChange: (inputStateCallback: (previousInputState: InputState) => InputState) => void,
    fieldValidators: FieldValidators
): ValidationPipeline {
    const { synchronousValidators = [], asynchronousValidators = [] } = fieldValidators

    function runSyncValidators(value: string): string | undefined {
        return head(compact(synchronousValidators.map(validator => validator(value))))
    }

    return function validationPipeline(events): Observable<ValidationResult> {
        const [changeEvents, blurEvents] = partition(events, event => event.type === 'change')
        let hasHadBlurError = false

        // Only emits when sync errors are found on blur
        const syncReasons: Observable<ValidationEvent> = blurEvents.pipe(
            map(event => event.target.value),
            filter(value => value.length > 0),
            map(runSyncValidators),
            filter(isDefined),
            map(reason => ({
                kind: 'BLUR' as const,
                reason,
            })),
            tap(() => {
                if (!hasHadBlurError) {
                    hasHadBlurError = true
                }
                onInputChange(previousInputState => ({ ...previousInputState, loading: false }))
            })
        )

        // Emits on every change event
        const changes: Observable<ValidationEvent> = changeEvents.pipe(
            map(event => event.target.value),
            map(value => {
                if (hasHadBlurError) {
                    const syncReason = runSyncValidators(value)
                    if (syncReason) {
                        return { kind: 'CHANGE_SYNC' as const, reason: syncReason, value }
                    }
                }

                return { kind: 'CHANGE_ASYNC' as const, value }
            }),
            tap(({ value, kind }) =>
                onInputChange(() => ({ value, loading: kind === 'CHANGE_ASYNC' && asynchronousValidators.length > 0 }))
            )
        )

        return merge(syncReasons, changes).pipe(
            debounce(validationEvent => timer(validationEvent.kind === 'CHANGE_ASYNC' ? typingDebounceTime : 0)),
            switchMap(validationEvent => {
                if (validationEvent.kind === 'BLUR') {
                    return of({ kind: 'INVALID' as const, reason: validationEvent.reason })
                }

                if (validationEvent.kind === 'CHANGE_SYNC') {
                    return of(
                        validationEvent.reason
                            ? { kind: 'INVALID' as const, reason: validationEvent.reason }
                            : { kind: 'VALID' as const }
                    )
                }

                return zip(...asynchronousValidators.map(validator => validator(validationEvent.value)))
                    .pipe(
                        map(values => head(compact(values))),
                        map(reason => (reason ? { kind: 'INVALID' as const, reason } : { kind: 'VALID' as const })),
                        tap(() => onInputChange(previousInputState => ({ ...previousInputState, loading: false })))
                    )
                    .pipe(takeUntil(syncReasons))
            }),
            catchError(() => of({ kind: 'INVALID' as const, reason: `Unknown error validating ${name}` }))
        )
    }
}

/**
 * The form for creating an account
 */
export const SignUpForm: React.FunctionComponent<SignUpFormProps> = ({ doSignUp, history, buttonLabel, className }) => {
    const [loading, setLoading] = useState(false)
    const [requestedTrial, setRequestedTrial] = useState(false)
    const [error, setError] = useState<Error | null>(null)

    const signUpFieldValidators: Record<'email' | 'username' | 'password', FieldValidators> = useMemo(
        () => ({
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
        }),
        []
    )

    const [emailState, setEmailState] = useState<InputState>({ value: '', loading: false })
    const [nextEmailFieldChange, emailValidationResult] = useEventObservable<
        React.ChangeEvent<HTMLInputElement>,
        ValidationResult
    >(
        useMemo(() => createValidationPipeline('email', setEmailState, signUpFieldValidators.email), [
            signUpFieldValidators,
        ])
    )

    const [usernameState, setUsernameState] = useState<InputState>({ value: '', loading: false })
    const [nextUsernameFieldChange, usernameValidationResult] = useEventObservable<
        React.ChangeEvent<HTMLInputElement>,
        ValidationResult
    >(
        useMemo(() => createValidationPipeline('username', setUsernameState, signUpFieldValidators.username), [
            signUpFieldValidators,
        ])
    )

    const [passwordState, setPasswordState] = useState<InputState>({ value: '', loading: false })
    const [nextPasswordFieldChange, passwordValidationResult] = useEventObservable<
        React.ChangeEvent<HTMLInputElement>,
        ValidationResult
    >(
        useMemo(() => createValidationPipeline('password', setPasswordState, signUpFieldValidators.password), [
            signUpFieldValidators,
        ])
    )

    const canRegister =
        emailValidationResult?.kind === 'VALID' &&
        usernameValidationResult?.kind === 'VALID' &&
        passwordValidationResult?.kind === 'VALID'

    const disabled = loading || !canRegister

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
                            onBlur={nextEmailFieldChange}
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
                        <small className="text-danger mt-2">{emailValidationResult.reason}</small>
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
                            onBlur={nextUsernameFieldChange}
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
                        <small className="text-danger mt-2">{usernameValidationResult.reason}</small>
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
                            onBlur={nextPasswordFieldChange}
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
                        <small className="text-danger mt-2">{passwordValidationResult.reason}</small>
                    ) : (
                        <small className="mt-2">At least 12 characters</small>
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
                        {loading ? <LoadingSpinner className="icon-inline" /> : buttonLabel || 'Register'}
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

function checkPasswordLength(password: string): string | undefined {
    if (password.length < 12) {
        return 'Password must contain at least 12 characters'
    }

    return undefined
}

function checkUsernameLength(username: string): string | undefined {
    if (username.length > USERNAME_MAX_LENGTH) {
        return `Username is longer than ${USERNAME_MAX_LENGTH} characters`
    }

    return undefined
}

function checkUsernamePattern(username: string): string | undefined {
    const valid = new RegExp(VALID_USERNAME_REGEXP).test(username)
    if (!valid) {
        return "Username doesn't match the requested format"
    }

    return undefined
}

/**
 * Simple email format validation to catch the most glaring errors
 * and display helpful error messages
 */
function checkEmailFormat(email: string): string | undefined {
    const parts = email.trim().split('@')
    if (parts.length < 2) {
        return "Please include an '@' in the email address"
    }
    if (parts.length > 2) {
        return "A part following '@' should not contain the symbol '@'"
    }

    return undefined
}

/**
 * Catch-all regex for errors not caught by `checkEmailFormat`.
 * From emailregex.com
 */
function checkEmailPattern(email: string): string | undefined {
    if (
        // eslint-disable-next-line no-useless-escape
        !/^(([^\s"(),.:;<>@[\\\]]+(\.[^\s"(),.:;<>@[\\\]]+)*)|(".+"))@((\[(?:\d{1,3}\.){3}\d{1,3}])|(([\dA-Za-z\-]+\.)+[A-Za-z]{2,}))$/.test(
            email
        )
    ) {
        return 'Please enter a valid email'
    }

    return undefined
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
