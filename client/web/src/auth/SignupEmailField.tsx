import React from 'react'

import classNames from 'classnames'

import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'
import { deriveInputClassName, InputValidationState } from '@sourcegraph/shared/src/util/useInputValidation'
import { Label } from '@sourcegraph/wildcard'

import { EmailInput } from './SignInSignUpCommon'

interface SignupEmailFieldProps {
    emailState: InputValidationState
    loading: boolean
    label: string
    nextEmailFieldChange: (changeEvent: React.ChangeEvent<HTMLInputElement>) => void
    emailInputReference: React.Ref<HTMLInputElement>
}

export const SignupEmailField: React.FunctionComponent<React.PropsWithChildren<SignupEmailFieldProps>> = ({
    emailState,
    loading,
    label,
    nextEmailFieldChange,
    emailInputReference,
}) => (
    <div className="form-group d-flex flex-column align-content-start">
        <Label
            htmlFor="email"
            className={classNames('align-self-start', {
                'text-danger font-weight-bold': emailState.kind === 'INVALID',
            })}
        >
            {label}
        </Label>
        <LoaderInput className={deriveInputClassName(emailState)} loading={emailState.kind === 'LOADING'}>
            <EmailInput
                className={deriveInputClassName(emailState)}
                onChange={nextEmailFieldChange}
                required={true}
                value={emailState.value}
                disabled={loading}
                autoFocus={true}
                placeholder=" "
                inputRef={emailInputReference}
                aria-describedby="email-input-invalid-feedback"
            />
        </LoaderInput>
        {emailState.kind === 'INVALID' && (
            <small className="invalid-feedback" id="email-input-invalid-feedback">
                {emailState.reason}
            </small>
        )}
    </div>
)
