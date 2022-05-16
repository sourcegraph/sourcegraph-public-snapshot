import React, { FunctionComponent, useMemo, useState } from 'react'

import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'
import { asError, isErrorLike, ErrorLike } from '@sourcegraph/common'
import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'
import { useInputValidation, deriveInputClassName } from '@sourcegraph/shared/src/util/useInputValidation'
import { screenReaderAnnounce } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../../backend/graphql'
import { LoaderButton } from '../../../components/LoaderButton'
import { AddUserEmailResult, AddUserEmailVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'

interface Props {
    user: string
    onDidAdd: () => void

    className?: string
}

type Status = undefined | 'loading' | ErrorLike

export const AddUserEmailForm: FunctionComponent<React.PropsWithChildren<Props>> = ({ user, className, onDidAdd }) => {
    const [statusOrError, setStatusOrError] = useState<Status>()

    const [emailState, nextEmailFieldChange, emailInputReference, overrideEmailState] = useInputValidation(
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

        if (emailState.kind === 'VALID') {
            setStatusOrError('loading')

            try {
                dataOrThrowErrors(
                    await requestGraphQL<AddUserEmailResult, AddUserEmailVariables>(
                        gql`
                            mutation AddUserEmail($user: ID!, $email: String!) {
                                addUserEmail(user: $user, email: $email) {
                                    alwaysNil
                                }
                            }
                        `,
                        { user, email: emailState.value }
                    ).toPromise()
                )

                eventLogger.log('NewUserEmailAddressAdded')
                screenReaderAnnounce('Email address added')

                overrideEmailState({ value: '' })
                setStatusOrError(undefined)

                if (onDidAdd) {
                    onDidAdd()
                }
            } catch (error) {
                setStatusOrError(asError(error))
            }
        }
    }

    return (
        <div className={classNames('add-user-email-form', className)}>
            <label
                htmlFor="AddUserEmailForm-email"
                className={classNames('align-self-start', {
                    'text-danger font-weight-bold': emailState.kind === 'INVALID',
                })}
            >
                Add email address
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
                    loading={statusOrError === 'loading'}
                    label="Add"
                    type="submit"
                    disabled={statusOrError === 'loading' || emailState.kind !== 'VALID'}
                    variant="primary"
                />
                {emailState.kind === 'INVALID' && (
                    <small className="invalid-feedback" role="alert">
                        {emailState.reason}
                    </small>
                )}
            </form>
            {isErrorLike(statusOrError) && <ErrorAlert className="mt-2" error={statusOrError} />}
        </div>
    )
}
