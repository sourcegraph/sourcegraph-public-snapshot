import React, { useState, FunctionComponent, useCallback } from 'react'
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
    error?: Error
}

export const SetUserPrimaryEmailForm: FunctionComponent<Props> = ({ user, emails, onDidSet, className, history }) => {
    const [primaryEmail, setPrimaryEmail] = useState<string>()
    const [status, setStatus] = useState<UserEmailState>({ loading: false })

    const options = emails.reduce((accumulator: string[], email) => {
        if (!email.isPrimary && email.verified) {
            accumulator.push(email.email)
        }
        return accumulator
    }, [])

    const defaultOption = emails.find(email => email.isPrimary)?.email
    const disabled = options.length === 0

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
                    defaultValue={defaultOption}
                >
                    {disabled ? (
                        <option value={defaultOption}>{defaultOption}</option>
                    ) : (
                        options.map(email => (
                            <option key={email} value={email}>
                                {email}
                            </option>
                        ))
                    )}
                </select>{' '}
                <LoaderButton
                    loading={status.loading}
                    label="Save"
                    type="submit"
                    disabled={disabled || status.loading}
                    className="btn btn-primary"
                />
            </Form>
            {status.error && <ErrorAlert className="mt-2" error={status.error} history={history} />}
        </div>
    )
}
