import React from 'react'

import { deriveInputClassName, type InputValidationState } from '@sourcegraph/shared/src/util/useInputValidation'
import { LoaderInput } from '@sourcegraph/wildcard'

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
    <div className="form-group">
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
                error={emailState.kind === 'INVALID' ? emailState.reason : undefined}
                label={label}
            />
        </LoaderInput>
    </div>
)
