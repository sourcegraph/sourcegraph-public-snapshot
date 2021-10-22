import classNames from 'classnames'
import cookies from 'js-cookie'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import React, { useCallback, useMemo, useState } from 'react'
import { Observable, of } from 'rxjs'
import { fromFetch } from 'rxjs/fetch'
import { catchError, switchMap } from 'rxjs/operators'

import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'
import { asError } from '@sourcegraph/shared/src/util/errors'
import {
    useInputValidation,
    ValidationOptions,
    deriveInputClassName,
} from '@sourcegraph/shared/src/util/useInputValidation'

import { ErrorAlert } from '../components/alerts'
import { LoaderButton } from '../components/LoaderButton'
import { FeatureFlagProps } from '../featureFlags/featureFlags'
import { AuthProvider, SourcegraphContext } from '../jscontext'
import { ANONYMOUS_USER_ID_KEY, eventLogger, FIRST_SOURCE_URL_KEY } from '../tracking/eventLogger'
import { enterpriseTrial, signupTerms } from '../util/features'

import { OrDivider } from './OrDivider'
import { maybeAddPostSignUpRedirect, PasswordInput, UsernameInput } from './SignInSignUpCommon'
import signInSignUpCommonStyles from './SignInSignUpCommon.module.scss'
import { SignupEmailField } from './SignupEmailField'
export interface SignUpArguments {
    email: string
    username: string
    password: string
    requestedTrial: boolean
    anonymousUserId?: string
    firstSourceUrl?: string
}

interface SignUpFormProps extends FeatureFlagProps {
    className?: string

    /** Called to perform the signup on the server. */
    onSignUp: (args: SignUpArguments) => Promise<void>

    buttonLabel?: string
    context: Pick<SourcegraphContext, 'authProviders' | 'sourcegraphDotComMode'>

    // For use in ExperimentalSignUpPage. Modifies styling and removes terms of service and trial section.
    experimental?: boolean
}

const preventDefault = (event: React.FormEvent): void => event.preventDefault()

/**
 * The form for creating an account
 */
export const SignUpForm: React.FunctionComponent<SignUpFormProps> = ({
    featureFlags,
    onSignUp,
    buttonLabel,
    className,
    context,
    experimental = false,
}) => {
    const [loading, setLoading] = useState(false)
    const [requestedTrial, setRequestedTrial] = useState(false)
    const [error, setError] = useState<Error | null>(null)
    const isSignupOptimised = featureFlags.get('signup-optimization')

    const signUpFieldValidators: Record<'email' | 'username' | 'password', ValidationOptions> = useMemo(
        () => ({
            email: {
                synchronousValidators: [],
                asynchronousValidators: [],
            },
            username: {
                synchronousValidators: [],
                asynchronousValidators: [isUsernameUnique],
            },
            password: {
                synchronousValidators: [],
            },
        }),
        []
    )

    const [emailState, nextEmailFieldChange, emailInputReference] = useInputValidation(signUpFieldValidators.email)

    const [usernameState, nextUsernameFieldChange, usernameInputReference] = useInputValidation(
        signUpFieldValidators.username
    )

    const [passwordState, nextPasswordFieldChange, passwordInputReference] = useInputValidation(
        signUpFieldValidators.password
    )

    const canRegister = emailState.kind === 'VALID' && usernameState.kind === 'VALID' && passwordState.kind === 'VALID'

    const disabled = loading || !canRegister

    const handleSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>): void => {
            event.preventDefault()
            if (disabled) {
                return
            }

            setLoading(true)
            onSignUp({
                email: emailState.value,
                username: usernameState.value,
                password: passwordState.value,
                requestedTrial,
                anonymousUserId: cookies.get(ANONYMOUS_USER_ID_KEY),
                firstSourceUrl: cookies.get(FIRST_SOURCE_URL_KEY),
            }).catch(error => {
                setError(asError(error))
                setLoading(false)
            })
            eventLogger.log('InitiateSignUp')
        },
        [onSignUp, disabled, emailState, usernameState, passwordState, requestedTrial]
    )

    const onRequestTrialFieldChange = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setRequestedTrial(event.target.checked)
    }, [])

    const externalAuthProviders = context.authProviders.filter(provider => !provider.isBuiltin)

    const onClickExternalAuthSignup = useCallback(
        (type: AuthProvider['serviceType']): React.MouseEventHandler<HTMLAnchorElement> => () => {
            // TODO: Log events with keepalive=true to ensure they always outlive the webpage
            // https://github.com/sourcegraph/sourcegraph/issues/19174
            eventLogger.log('SignupInitiated', { type }, { type })
        },
        []
    )
    return (
        <>
            {error && <ErrorAlert className="mt-4 mb-0" error={error} />}
            {/* Using  <form /> to set 'valid' + 'is-invaild' at the input level */}
            {/* eslint-disable-next-line react/forbid-elements */}
            <form
                className={classNames(
                    !experimental && signInSignUpCommonStyles.signinSignupForm,
                    'test-signup-form',
                    !experimental && 'rounded p-4',
                    'text-left',
                    !experimental && (context.sourcegraphDotComMode || error) ? 'mt-3' : 'mt-4',
                    className
                )}
                onSubmit={handleSubmit}
                noValidate={true}
            >
                {!isSignupOptimised && (
                    <SignupEmailField
                        label="Email"
                        loading={loading}
                        nextEmailFieldChange={nextEmailFieldChange}
                        emailState={emailState}
                        emailInputReference={emailInputReference}
                    />
                )}
                <div className="form-group d-flex flex-column align-content-start">
                    <label
                        htmlFor="username"
                        className={classNames('align-self-start', {
                            'text-danger font-weight-bold': usernameState.kind === 'INVALID',
                        })}
                    >
                        Username
                    </label>
                    <LoaderInput
                        className={classNames(deriveInputClassName(usernameState))}
                        loading={usernameState.kind === 'LOADING'}
                    >
                        <UsernameInput
                            className={deriveInputClassName(usernameState)}
                            onChange={nextUsernameFieldChange}
                            value={usernameState.value}
                            required={true}
                            disabled={loading}
                            placeholder=" "
                            inputRef={usernameInputReference}
                        />
                    </LoaderInput>
                    {usernameState.kind === 'INVALID' && (
                        <small className="invalid-feedback" role="alert">
                            {usernameState.reason}
                        </small>
                    )}
                </div>
                {isSignupOptimised && (
                    <SignupEmailField
                        label="Email address"
                        loading={loading}
                        nextEmailFieldChange={nextEmailFieldChange}
                        emailState={emailState}
                        emailInputReference={emailInputReference}
                    />
                )}
                <div className="form-group d-flex flex-column align-content-start">
                    <label
                        htmlFor="password"
                        className={classNames('align-self-start', {
                            'text-danger font-weight-bold': passwordState.kind === 'INVALID',
                        })}
                    >
                        Password
                    </label>
                    <LoaderInput
                        className={classNames(deriveInputClassName(passwordState))}
                        loading={passwordState.kind === 'LOADING'}
                    >
                        <PasswordInput
                            className={deriveInputClassName(passwordState)}
                            onChange={nextPasswordFieldChange}
                            value={passwordState.value}
                            required={true}
                            disabled={loading}
                            autoComplete="new-password"
                            placeholder=" "
                            onInvalid={preventDefault}
                            minLength={12}
                            inputRef={passwordInputReference}
                            formNoValidate={true}
                        />
                    </LoaderInput>
                    {passwordState.kind === 'INVALID' ? (
                        <small className="invalid-feedback" role="alert">
                            {passwordState.reason}
                        </small>
                    ) : (
                        <small className="form-text text-muted">At least 12 characters</small>
                    )}
                </div>
                {!experimental && enterpriseTrial && (
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
                    <LoaderButton
                        loading={loading}
                        label={buttonLabel || 'Register'}
                        type="submit"
                        disabled={disabled}
                        className={classNames('btn btn-primary btn-block', isSignupOptimised && 'mt-4')}
                    />
                </div>
                {context.sourcegraphDotComMode && (
                    <>
                        {externalAuthProviders.length > 0 && <OrDivider className="my-4" />}
                        {externalAuthProviders.map((provider, index) => (
                            // Use index as key because display name may not be unique. This is OK
                            // here because this list will not be updated during this component's lifetime.
                            <div className="mb-2" key={index}>
                                <a
                                    href={maybeAddPostSignUpRedirect(provider.authenticationURL)}
                                    className="btn btn-secondary btn-block"
                                    onClick={onClickExternalAuthSignup(provider.serviceType)}
                                >
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

                {!experimental && signupTerms && (
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
            </form>
        </>
    )
}

// Asynchronous Validators

function isUsernameUnique(username: string): Observable<string | undefined> {
    return fromFetch(`/-/check-username-taken/${username}`).pipe(
        switchMap(response => {
            switch (response.status) {
                case 200:
                    return of('Username is already taken.')
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
