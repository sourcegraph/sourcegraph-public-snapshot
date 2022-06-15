import React, { useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'
import cookies from 'js-cookie'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import { Observable, of } from 'rxjs'
import { fromFetch } from 'rxjs/fetch'
import { catchError, switchMap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'
import { asError } from '@sourcegraph/common'
import {
    useInputValidation,
    ValidationOptions,
    deriveInputClassName,
} from '@sourcegraph/shared/src/util/useInputValidation'
import { Button, Link, Icon, Checkbox, Label, Text } from '@sourcegraph/wildcard'

import { LoaderButton } from '../components/LoaderButton'
import { AuthProvider, SourcegraphContext } from '../jscontext'
import { ANONYMOUS_USER_ID_KEY, eventLogger, FIRST_SOURCE_URL_KEY, LAST_SOURCE_URL_KEY } from '../tracking/eventLogger'
import { enterpriseTrial } from '../util/features'

import { OrDivider } from './OrDivider'
import { maybeAddPostSignUpRedirect, PasswordInput, UsernameInput } from './SignInSignUpCommon'
import { SignupEmailField } from './SignupEmailField'

import signInSignUpCommonStyles from './SignInSignUpCommon.module.scss'

export interface SignUpArguments {
    email: string
    username: string
    password: string
    requestedTrial: boolean
    anonymousUserId?: string
    firstSourceUrl?: string
    lastSourceUrl?: string
}

interface SignUpFormProps {
    className?: string

    /** Called to perform the signup on the server. */
    onSignUp: (args: SignUpArguments) => Promise<void>

    buttonLabel?: string
    context: Pick<SourcegraphContext, 'authProviders' | 'sourcegraphDotComMode' | 'experimentalFeatures'>

    // For use in ExperimentalSignUpPage. Modifies styling and removes terms of service and trial section.
    experimental?: boolean
}

const preventDefault = (event: React.FormEvent): void => event.preventDefault()

export function getPasswordRequirements(
    context: Pick<SourcegraphContext, 'authProviders' | 'sourcegraphDotComMode' | 'experimentalFeatures'>
): string {
    let requirements = ''
    const passwordPolicyReference = context.experimentalFeatures.passwordPolicy

    if (passwordPolicyReference && passwordPolicyReference.enabled === true) {
        console.log('Using enhanced password policy.')

        if (passwordPolicyReference.minimumLength && passwordPolicyReference.minimumLength > 0) {
            requirements +=
                'Your password must include at least ' + String(passwordPolicyReference.minimumLength) + ' characters'
        }
        if (
            passwordPolicyReference.numberOfSpecialCharacters &&
            passwordPolicyReference.numberOfSpecialCharacters > 0
        ) {
            requirements += ', ' + String(passwordPolicyReference.numberOfSpecialCharacters) + ' special characters'
        }
        if (
            passwordPolicyReference.requireAtLeastOneNumber &&
            passwordPolicyReference.requireAtLeastOneNumber === true
        ) {
            requirements += ', at least one number'
        }
        if (
            passwordPolicyReference.requireUpperandLowerCase &&
            passwordPolicyReference.requireUpperandLowerCase === true
        ) {
            requirements += ', at least one uppercase letter'
        }
    } else {
        requirements += 'At least 12 characters'
    }
    return requirements
}

/**
 * The form for creating an account
 */
export const SignUpForm: React.FunctionComponent<React.PropsWithChildren<SignUpFormProps>> = ({
    onSignUp,
    buttonLabel,
    className,
    context,
    experimental = false,
}) => {
    const [loading, setLoading] = useState(false)
    const [requestedTrial, setRequestedTrial] = useState(false)
    const [error, setError] = useState<Error | null>(null)

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
                synchronousValidators: [password => validatePassword(context, password)],
                asynchronousValidators: [],
            },
        }),
        [context]
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
                lastSourceUrl: cookies.get(LAST_SOURCE_URL_KEY),
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
        (type: AuthProvider['serviceType']): React.MouseEventHandler<HTMLButtonElement> => () => {
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
                <SignupEmailField
                    label="Email"
                    loading={loading}
                    nextEmailFieldChange={nextEmailFieldChange}
                    emailState={emailState}
                    emailInputReference={emailInputReference}
                />
                <div className="form-group d-flex flex-column align-content-start">
                    <Label
                        htmlFor="username"
                        className={classNames('align-self-start', {
                            'text-danger font-weight-bold': usernameState.kind === 'INVALID',
                        })}
                    >
                        Username
                    </Label>
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
                            aria-describedby="username-input-invalid-feedback"
                        />
                    </LoaderInput>
                    {usernameState.kind === 'INVALID' && (
                        <small className="invalid-feedback" id="username-input-invalid-feedback" role="alert">
                            {usernameState.reason}
                        </small>
                    )}
                </div>
                <div className="form-group d-flex flex-column align-content-start">
                    <Label
                        htmlFor="password"
                        className={classNames('align-self-start', {
                            'text-danger font-weight-bold': passwordState.kind === 'INVALID',
                        })}
                    >
                        Password
                    </Label>
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
                            minLength={
                                context.experimentalFeatures.passwordPolicy?.enabled !== undefined &&
                                context.experimentalFeatures.passwordPolicy.enabled &&
                                context.experimentalFeatures.passwordPolicy?.minimumLength !== undefined
                                    ? context.experimentalFeatures.passwordPolicy.minimumLength
                                    : 12
                            }
                            placeholder=" "
                            onInvalid={preventDefault}
                            inputRef={passwordInputReference}
                            formNoValidate={true}
                            aria-describedby="password-input-invalid-feedback password-requirements"
                        />
                    </LoaderInput>
                    {passwordState.kind === 'INVALID' && (
                        <small className="invalid-feedback" id="password-input-invalid-feedback" role="alert">
                            {passwordState.reason}
                        </small>
                    )}
                    <small className="form-help text-muted" id="password-requirements">
                        {getPasswordRequirements(context)}
                    </small>
                </div>
                {!experimental && enterpriseTrial && (
                    <div className="form-group">
                        <Checkbox
                            onChange={onRequestTrialFieldChange}
                            id="EnterpriseTrialCheck"
                            label={
                                <>
                                    Try Sourcegraph Enterprise free for
                                    <span className="text-nowrap">
                                        30 days{' '}
                                        <Link target="_blank" rel="noopener" to="https://about.sourcegraph.com/pricing">
                                            <Icon as={HelpCircleOutlineIcon} />
                                        </Link>
                                    </span>
                                </>
                            }
                        />
                    </div>
                )}
                <div className="form-group mb-0">
                    <LoaderButton
                        loading={loading}
                        label={buttonLabel || 'Register'}
                        type="submit"
                        disabled={disabled}
                        className="btn-block"
                        variant="primary"
                    />
                </div>
                {context.sourcegraphDotComMode && (
                    <>
                        {externalAuthProviders.length > 0 && <OrDivider className="my-4" />}
                        {externalAuthProviders.map((provider, index) => (
                            // Use index as key because display name may not be unique. This is OK
                            // here because this list will not be updated during this component's lifetime.
                            <div className="mb-2" key={index}>
                                <Button
                                    href={maybeAddPostSignUpRedirect(provider.authenticationURL)}
                                    className="btn-block"
                                    onClick={onClickExternalAuthSignup(provider.serviceType)}
                                    variant="secondary"
                                    as="a"
                                >
                                    {provider.serviceType === 'github' ? (
                                        <Icon as={GithubIcon} />
                                    ) : provider.serviceType === 'gitlab' ? (
                                        <Icon as={GitlabIcon} />
                                    ) : null}{' '}
                                    Continue with {provider.displayName}
                                </Button>
                            </div>
                        ))}
                    </>
                )}

                {!experimental && (
                    <Text className="mt-3 mb-0">
                        <small className="form-text text-muted">
                            By signing up, you agree to our{' '}
                            <Link to="https://about.sourcegraph.com/terms" target="_blank" rel="noopener">
                                Terms of Service
                            </Link>{' '}
                            and{' '}
                            <Link to="https://about.sourcegraph.com/privacy" target="_blank" rel="noopener">
                                Privacy Policy
                            </Link>
                            .
                        </small>
                    </Text>
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

function validatePassword(
    context: Pick<SourcegraphContext, 'authProviders' | 'sourcegraphDotComMode' | 'experimentalFeatures'>,
    password: string
): string | undefined {
    if (context.experimentalFeatures.passwordPolicy?.enabled) {
        if (
            context.experimentalFeatures.passwordPolicy.minimumLength &&
            password.length < context.experimentalFeatures.passwordPolicy.minimumLength
        ) {
            return (
                'Password must be greater than ' +
                String(context.experimentalFeatures.passwordPolicy.minimumLength) +
                ' characters.'
            )
        }
        if (
            context.experimentalFeatures.passwordPolicy?.numberOfSpecialCharacters &&
            context.experimentalFeatures.passwordPolicy.numberOfSpecialCharacters > 0
        ) {
            const specialCharacters = /[!"#$%&'()*+,./:;<=>?@[\]^_`{|}~-]/
            // This must be kept in sync with the security.go checks
            const count = (password.match(specialCharacters) || []).length
            if (
                context.experimentalFeatures.passwordPolicy.numberOfSpecialCharacters &&
                count < context.experimentalFeatures.passwordPolicy.numberOfSpecialCharacters
            ) {
                return (
                    'Password must contain ' +
                    String(context.experimentalFeatures.passwordPolicy.numberOfSpecialCharacters) +
                    ' special character(s).'
                )
            }
        }

        if (
            context.experimentalFeatures.passwordPolicy.requireAtLeastOneNumber &&
            context.experimentalFeatures.passwordPolicy.requireAtLeastOneNumber
        ) {
            const validRequireAtLeastOneNumber = /\d+/
            if (password.match(validRequireAtLeastOneNumber) === null) {
                return 'Password must contain at least one number.'
            }
        }

        if (
            context.experimentalFeatures.passwordPolicy.requireUpperandLowerCase &&
            context.experimentalFeatures.passwordPolicy.requireUpperandLowerCase
        ) {
            const validUseUpperCase = new RegExp('[A-Z]+')
            if (!validUseUpperCase.test(password)) {
                return 'Password must contain at least one uppercase letter.'
            }
        }

        return undefined
    }

    if (password.length < 12) {
        return 'Password must be at least 12 characters.'
    }

    return undefined
}
