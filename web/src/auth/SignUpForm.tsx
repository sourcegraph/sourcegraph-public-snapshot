import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import React, { useCallback, useMemo, useState, useRef } from 'react'
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
import { catchError, debounce, debounceTime, filter, map, switchMap, tap } from 'rxjs/operators'
import { typingDebounceTime } from '../search/input/QueryInput'
import { fromFetch } from 'rxjs/fetch'
import { isDefined } from '../../../shared/src/util/types'
import GitlabIcon from 'mdi-react/GitlabIcon'
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
    fieldValidators: FieldValidators,
    inputReference: React.MutableRefObject<HTMLInputElement | null>
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
                inputReference.current?.setCustomValidity('')
                return zip(...asynchronousValidators.map(validator => validator(validationEvent.value))).pipe(
                    map(values => head(compact(values))),
                    map(reason => (reason ? { kind: 'INVALID' as const, reason } : { kind: 'VALID' as const })),
                    tap(({ reason }) => {
                        if (reason) {
                            inputReference.current?.setCustomValidity(reason)
                        }
                        onInputChange(previousInputState => ({ ...previousInputState, loading: false }))
                    })
                )
            }),
            catchError(() => of({ kind: 'INVALID' as const, reason: `Unknown error validating ${name}` }))
        )
    }
}

function createSimpleValidationPipeline(
    name: string,
    onInputChange: (inputStateCallback: (previousInputState: InputState) => InputState) => void,
    fieldValidators: FieldValidators,
    inputReference: React.MutableRefObject<HTMLInputElement | null>
): ValidationPipeline {
    const { synchronousValidators = [], asynchronousValidators = [] } = fieldValidators

    return function validationPipeline(
        events: Observable<React.ChangeEvent<HTMLInputElement>>
    ): Observable<ValidationResult> {
        // debounce everything
        return events.pipe(
            tap(event => {
                event.preventDefault()
                // capture sync messages, the set custom validation to "" for loading neutral state
            }),
            map(event => event.target.value),
            tap(value => onInputChange(() => ({ value, loading: true }))),
            debounceTime(typingDebounceTime),
            switchMap(value => {
                // check validity (synchronous)
                const valid = inputReference.current?.checkValidity()
                console.log('value: ' + value + ' is: ', valid)
                if (!valid) {
                    // TODO: BUG
                    console.log('the validation message is:', inputReference.current?.validationMessage)

                    // inputReference.current?.setCustomValidity
                    return of({ kind: 'INVALID' as const, reason: inputReference.current?.validationMessage ?? '' })
                }

                // check any custom sync validators
                const syncReason = head(compact(synchronousValidators.map(validator => validator(value))))
                if (syncReason) {
                    inputReference.current?.setCustomValidity(syncReason)
                    return of({ kind: 'INVALID' as const, reason: syncReason })
                }

                // check async validators
                return zip(...asynchronousValidators.map(validator => validator(value))).pipe(
                    map(values => head(compact(values))),
                    map(reason => (reason ? { kind: 'INVALID' as const, reason } : { kind: 'VALID' as const })),
                    tap(result => {
                        if (result.kind === 'INVALID') {
                            inputReference.current?.setCustomValidity(result.reason)
                        } else {
                            inputReference.current?.setCustomValidity('')
                        }
                        onInputChange(previousInputState => ({ ...previousInputState, loading: false }))
                    })
                )
            }),
            tap(() => onInputChange(previousInputState => ({ ...previousInputState, loading: false }))),
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
                synchronousValidators: [],
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

    const emailInputReference = useRef<HTMLInputElement | null>(null)

    const [emailState, setEmailState] = useState<InputState>({ value: '', loading: false })
    const [nextEmailFieldChange, emailValidationResult] = useEventObservable<
        React.ChangeEvent<HTMLInputElement>,
        ValidationResult
    >(
        useMemo(
            () =>
                createSimpleValidationPipeline(
                    'email',
                    setEmailState,
                    signUpFieldValidators.email,
                    emailInputReference
                ),
            [signUpFieldValidators]
        )
    )

    const usernameInputReference = useRef<HTMLInputElement | null>(null)

    const [usernameState, setUsernameState] = useState<InputState>({ value: '', loading: false })
    const [nextUsernameFieldChange, usernameValidationResult] = useEventObservable<
        React.ChangeEvent<HTMLInputElement>,
        ValidationResult
    >(
        useMemo(
            () =>
                createValidationPipeline(
                    'username',
                    setUsernameState,
                    signUpFieldValidators.username,
                    usernameInputReference
                ),
            [signUpFieldValidators]
        )
    )

    const passwordInputReference = useRef<HTMLInputElement | null>(null)

    const [passwordState, setPasswordState] = useState<InputState>({ value: '', loading: false })
    const [nextPasswordFieldChange, passwordValidationResult] = useEventObservable<
        React.ChangeEvent<HTMLInputElement>,
        ValidationResult
    >(
        useMemo(
            () =>
                createValidationPipeline(
                    'password',
                    setPasswordState,
                    signUpFieldValidators.password,
                    passwordInputReference
                ),
            [signUpFieldValidators]
        )
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

    const preventDefault = useCallback((event: React.FormEvent) => event.preventDefault(), [])

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
                // noValidate={true}
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
                    <div
                        className={classNames(
                            'signin-signup-form__input-container',
                            emailInputReference.current?.validationMessage && 'is-invalid'
                        )}
                    >
                        <EmailInput
                            className="signin-signup-form__input"
                            onChange={nextEmailFieldChange}
                            onBlur={nextEmailFieldChange}
                            required={true}
                            value={emailState.value}
                            disabled={loading}
                            autoFocus={true}
                            placeholder=" "
                            inputRef={emailInputReference}
                        />
                        {emailState.loading && <LoadingSpinner className="signin-signup-form__icon" />}
                    </div>
                    {!emailState.loading && emailValidationResult?.kind === 'INVALID' && (
                        <small className="invalid-feedback">{emailValidationResult.reason}</small>
                    )}
                </div>
                <div className="form-group d-flex flex-column align-content-start">
                    <label
                        className={classNames('align-self-start', {
                            'text-danger font-weight-bold':
                                !usernameInputReference.current?.validity.valid && !usernameState.loading,
                        })}
                    >
                        Username
                    </label>
                    <div
                        className={classNames(
                            'signin-signup-form__input-container',
                            !usernameInputReference.current?.validity.valid && 'is-invalid'
                        )}
                    >
                        <UsernameInput
                            className="signin-signup-form__input"
                            onChange={nextUsernameFieldChange}
                            onBlur={nextUsernameFieldChange}
                            value={usernameState.value}
                            required={true}
                            disabled={loading}
                            placeholder=" "
                            pattern={VALID_USERNAME_REGEXP}
                        />
                        {usernameState.loading && <LoadingSpinner className="signin-signup-form__icon" />}
                    </div>
                    {!usernameState.loading && !usernameInputReference.current?.validity.valid && (
                        <small className="invalid-feedback" role="alert">
                            {usernameInputReference.current?.validationMessage}
                        </small>
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
                    <div
                        className={classNames(
                            'signin-signup-form__input-container',
                            passwordInputReference.current?.validationMessage && 'is-invalid'
                        )}
                    >
                        <PasswordInput
                            className="signin-signup-form__input"
                            onChange={nextPasswordFieldChange}
                            onBlur={nextPasswordFieldChange}
                            value={passwordState.value}
                            required={true}
                            disabled={loading}
                            autoComplete="new-password"
                            placeholder=" "
                            onInvalid={preventDefault}
                            minLength={12}
                        />
                        {passwordState.loading && <LoadingSpinner className="signin-signup-form__icon" />}
                    </div>
                    {!passwordState.loading && passwordInputReference.current?.validationMessage ? (
                        <small className="invalid-feedback" role="alert">
                            {passwordInputReference.current?.validationMessage}
                        </small>
                    ) : (
                        <small className="form-text">At least 12 characters</small>
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
                                    {provider.serviceType === 'github' ? (
                                        <GithubIcon className="icon-inline" />
                                    ) : provider.serviceType === 'gitlab' ? (
                                        <GitlabIcon className="icon-inline" />
                                    ) : null}{' '}
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
