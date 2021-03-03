import React, { FunctionComponent, useMemo } from 'react'
import classNames from 'classnames'
import * as H from 'history'

import { AddUserEmailResult, AddUserEmailVariables } from '../../../graphql-operations'
import { useInputValidation, deriveInputClassName } from '../../../../../shared/src/util/useInputValidation'

import { eventLogger } from '../../../tracking/eventLogger'
import { ErrorAlert } from '../../../components/alerts'
import { LoaderButton } from '../../../components/LoaderButton'
import { LoaderInput } from '../../../../../branded/src/components/LoaderInput'
import { gql, useMutation } from '@apollo/client'
import { FETCH_USER_EMAILS } from './UserSettingsEmailsPage'

interface Props {
    user: string
    history: H.History

    className?: string
}

const ADD_USER_EMAIL = gql`
    mutation AddUserEmail($user: ID!, $email: String!) {
        addUserEmail(user: $user, email: $email) {
            alwaysNil
        }
    }
`

export const AddUserEmailForm: FunctionComponent<Props> = ({ user, className, history }) => {
    const [addUserEmail, { error, loading }] = useMutation<AddUserEmailResult, AddUserEmailVariables>(ADD_USER_EMAIL, {
        onCompleted: () => eventLogger.log('NewUserEmailAddressAdded'),
        refetchQueries: [{ query: FETCH_USER_EMAILS, variables: { user } }],
    })

    const [emailState, nextEmailFieldChange, emailInputReference] = useInputValidation(
        useMemo(
            () => ({
                synchronousValidators: [],
                asynchronousValidators: [],
            }),
            []
        )
    )

    const onSubmit: React.FormEventHandler<HTMLFormElement> = async event => {
        event.preventDefault()
        return addUserEmail({ variables: { user, email: emailState.value } })
    }

    return (
        <div className={`add-user-email-form ${className || ''}`}>
            <label
                htmlFor="AddUserEmailForm-email"
                className={classNames('align-self-start', {
                    'text-danger font-weight-bold': emailState.kind === 'INVALID',
                })}
            >
                Email address
            </label>
            {/* eslint-disable-next-line react/forbid-elements */}
            <form className="form-inline" onSubmit={onSubmit} noValidate={true}>
                <LoaderInput
                    className={classNames(deriveInputClassName(emailState), 'mr-sm-2')}
                    loading={emailState.kind === 'LOADING'}
                >
                    <input
                        id="AddUserEmailForm-email"
                        type="email"
                        name="email"
                        className={classNames(
                            'form-control test-user-email-add-input',
                            deriveInputClassName(emailState)
                        )}
                        onChange={nextEmailFieldChange}
                        size={32}
                        value={emailState.value}
                        ref={emailInputReference}
                        required={true}
                        autoComplete="email"
                        autoCorrect="off"
                        autoCapitalize="off"
                        spellCheck={false}
                        readOnly={false}
                    />
                </LoaderInput>
                <LoaderButton
                    loading={loading}
                    label="Add"
                    type="submit"
                    disabled={loading || emailState.kind !== 'VALID'}
                    className="btn btn-primary"
                />
                {emailState.kind === 'INVALID' && (
                    <small className="invalid-feedback" role="alert">
                        {emailState.reason}
                    </small>
                )}
            </form>
            {error && <ErrorAlert className="mt-2" error={error} history={history} />}
        </div>
    )
}
