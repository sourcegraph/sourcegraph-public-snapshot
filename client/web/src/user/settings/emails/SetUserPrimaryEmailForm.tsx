import React, { useState, FunctionComponent, useCallback } from 'react'
import * as H from 'history'

import { SetUserEmailPrimaryResult, SetUserEmailPrimaryVariables, UserEmailsResult } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'

import { Form } from '../../../../../branded/src/components/Form'
import { LoaderButton } from '../../../components/LoaderButton'
import { ErrorAlert } from '../../../components/alerts'
import { gql, useMutation } from '@apollo/client'
import { FETCH_USER_EMAILS } from './UserSettingsEmailsPage'

type UserEmail = NonNullable<UserEmailsResult['node']>['emails'][number]

interface Props {
    user: string
    emails: UserEmail[]
    history: H.History

    className?: string
}

// There is always exactly one primary email returned from the backend
// eslint-disable-next-line @typescript-eslint/no-non-null-assertion
const findPrimaryEmail = (emails: UserEmail[]): string => emails.find(email => email.isPrimary)!.email

const SET_USER_EMAIL_PRIMARY = gql`
    mutation SetUserEmailPrimary($user: ID!, $email: String!) {
        setUserEmailPrimary(user: $user, email: $email) {
            alwaysNil
        }
    }
`

export const SetUserPrimaryEmailForm: FunctionComponent<Props> = ({ user, emails, className, history }) => {
    const [updatePrimaryUserEmail, { loading, error }] = useMutation<
        SetUserEmailPrimaryResult,
        SetUserEmailPrimaryVariables
    >(SET_USER_EMAIL_PRIMARY, {
        onCompleted: () => eventLogger.log('UserEmailAddressSetAsPrimary'),
        refetchQueries: [{ query: FETCH_USER_EMAILS, variables: { user } }],
    })

    const [primaryEmail, setPrimaryEmail] = useState<string>(findPrimaryEmail(emails))

    // options should include all verified emails + a primary one
    const options = emails.filter(email => email.verified || email.isPrimary).map(email => email.email)

    const onPrimaryEmailSelect: React.ChangeEventHandler<HTMLSelectElement> = event =>
        setPrimaryEmail(event.target.value)

    const onSubmit: React.FormEventHandler<HTMLFormElement> = useCallback(
        event => {
            event.preventDefault()
            return updatePrimaryUserEmail({ variables: { user, email: primaryEmail } })
        },
        [user, primaryEmail, updatePrimaryUserEmail]
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
                    disabled={options.length === 1 || loading}
                >
                    {options.map(email => (
                        <option key={email} value={email}>
                            {email}
                        </option>
                    ))}
                </select>
                <LoaderButton
                    loading={loading}
                    label="Save"
                    type="submit"
                    disabled={options.length === 1 || loading}
                    className="btn btn-primary"
                />
            </Form>
            {error && <ErrorAlert className="mt-2" error={error} history={history} />}
        </div>
    )
}
