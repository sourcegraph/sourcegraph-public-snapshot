import React, { useState, FunctionComponent, useEffect, useCallback } from 'react'
import * as H from 'history'

import { IUserEmail } from '../../../../../shared/src/graphql/schema'
import { requestGraphQL } from '../../../backend/graphql'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import { eventLogger } from '../../../tracking/eventLogger'

import { Form } from '../../../../../branded/src/components/Form'
import { LoaderButton } from '../../../components/LoaderButton'
import { ErrorAlert } from '../../../components/alerts'
import { SetUserEmailPrimaryResult, SetUserEmailPrimaryVariables } from '../../../graphql-operations'

interface Props {
    user: string
    emails: Omit<IUserEmail, '__typename' | 'user'>[]
    onDidSet: (email: string) => void
    history: H.History
    className?: string
}

interface UserEmailState {
    loading: boolean
    errorDescription?: Error
}

export const SetUserPrimaryEmailForm: FunctionComponent<Props> = ({ user, emails, onDidSet, className, history }) => {
    const [primaryEmail, setPrimaryEmail] = useState<string>('')
    const [emailOptions, setEmailOptions] = useState<string[]>([])
    const [status, setStatus] = useState<UserEmailState>({ loading: false })
    const [disabled, setDisabled] = useState(false)

    useEffect(() => {
        const possibleEmails = []
        let placeholderPrimaryEmail = ''

        /**
         * Every time emails props changes, find:
         *
         * 1. all emails that can be set as primary (not primary and verified)
         * 2. current primary email to be used as UI "placeholder" for disabled
         * select
         */
        for (const email of emails) {
            // collect possible primary emails
            if (!email.isPrimary && email.verified) {
                possibleEmails.push(email.email)
            }

            if (email.isPrimary) {
                // there can be only one
                placeholderPrimaryEmail = email.email
            }
        }

        const shouldDisable = possibleEmails.length === 0

        let options
        let selectValue

        /**
         * If possible primary emails were found, use them as select's options
         * and set the first email as select's value
         */
        if (possibleEmails.length !== 0) {
            options = possibleEmails
            selectValue = possibleEmails[0]
        } else {
            options = [placeholderPrimaryEmail]
            selectValue = placeholderPrimaryEmail
        }

        setDisabled(shouldDisable)
        setEmailOptions(options)
        setPrimaryEmail(selectValue)
    }, [emails])

    const onPrimaryEmailSelect: React.ChangeEventHandler<HTMLSelectElement> = event =>
        setPrimaryEmail(event.target.value)

    const onSubmit: React.FormEventHandler<HTMLFormElement> = useCallback(
        async event => {
            event.preventDefault()
            setStatus({ loading: true })

            try {
                dataOrThrowErrors(
                    await requestGraphQL<SetUserEmailPrimaryResult, SetUserEmailPrimaryVariables>(
                        gql`
                            mutation SetUserEmailPrimary($user: ID!, $email: String!) {
                                setUserEmailPrimary(user: $user, email: $email) {
                                    alwaysNil
                                }
                            }
                        `,
                        { user, email: primaryEmail }
                    ).toPromise()
                )
            } catch (error) {
                setStatus({ loading: false, errorDescription: error as Error })
                return
            }

            eventLogger.log('UserEmailAddressSetAsPrimary')
            setStatus({ loading: false })

            if (onDidSet) {
                onDidSet(primaryEmail)
            }
        },
        [user, primaryEmail, onDidSet]
    )

    return (
        <div className={`add-user-email-form ${className || ''}`}>
            <p className="mb-2">Primary email address</p>
            <Form className="form-inline" onSubmit={onSubmit}>
                <label className="sr-only" htmlFor="setUserPrimaryEmailForm-email">
                    Email address
                </label>
                <select
                    id="setUserPrimaryEmailForm-email"
                    className="custom-select form-control-lg mr-sm-2"
                    value={primaryEmail}
                    onChange={onPrimaryEmailSelect}
                    required={true}
                    disabled={disabled}
                >
                    {emailOptions.map(email => (
                        <option key={email} value={email}>
                            {email}
                        </option>
                    ))}
                </select>{' '}
                <LoaderButton
                    loading={status.loading}
                    label="Save"
                    type="submit"
                    disabled={disabled || status.loading}
                    className="btn btn-primary"
                />
            </Form>
            {status.errorDescription && (
                <ErrorAlert className="mt-2" error={status.errorDescription} history={history} />
            )}
        </div>
    )
}
