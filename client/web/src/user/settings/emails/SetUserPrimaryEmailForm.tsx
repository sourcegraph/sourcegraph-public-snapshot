/* eslint-disable react/jsx-no-bind */
import React, { useState, FunctionComponent } from 'react'
import * as H from 'history'
import { Form } from '../../../../../branded/src/components/Form'
import { IUserEmail } from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'
import { gql } from '../../../../../shared/src/graphql/graphql'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { eventLogger } from '../../../tracking/eventLogger'
import { ErrorAlert } from '../../../components/alerts'

interface Props {
    userId: string
    emails: IUserEmail[]
    onDidSet: (email: string) => void
    history: H.History
    className?: string
}

interface UserEmailState {
    loading: boolean
    errorDescription: Error | null
}

export const SetUserPrimaryEmailForm: FunctionComponent<Props> = ({ userId, emails, onDidSet, className, history }) => {
    const [primaryEmail, setPrimaryEmail] = useState('')
    const [status, setStatus] = useState<UserEmailState>({ loading: false, errorDescription: null })

    const onPrimaryEmailSelect: React.ChangeEventHandler<HTMLSelectElement> = event =>
        setPrimaryEmail(event.target.value)

    const onSubmit: React.FormEventHandler<HTMLFormElement> = async event => {
        event.preventDefault()

        const { data, errors } = mutateGraphQL(
            gql`
                mutation SetUserEmailPrimary($user: ID!, $email: String!) {
                    setUserEmailPrimary(user: $user, email: $email) {
                        alwaysNil
                    }
                }
            `,
            { user: userId, email: primaryEmail }
        ).toPromise()

        // TODO: check this
        if (!data || (errors && errors.length > 0)) {
            const aggregateError = createAggregateError(errors)
            setStatus({ loading: false, errorDescription: aggregateError })
            throw aggregateError
        }

        setStatus({ ...status, loading: false })

        if (onDidSet) {
            onDidSet(primaryEmail)
        }
        eventLogger.log('UserEmailAddressSetAsPrimary')
    }

    return (
        <div className={`add-user-email-form ${className || ''}`}>
            <h3>Set primary email address</h3>
            <Form className="form-inline" onSubmit={onSubmit}>
                <label className="sr-only" htmlFor="setUserPrimaryEmailForm-email">
                    Email address
                </label>
                <select
                    id="setUserPrimaryEmailForm-email"
                    className="custom-select form-control mr-sm-2"
                    value={primaryEmail}
                    onChange={onPrimaryEmailSelect}
                    required={true}
                    disabled={true}
                >
                    {emails.reduce((options, email) => {
                        if (!email.isPrimary /* && email.verified*/) {
                            options.push(
                                <option key={email.email} value={email.email}>
                                    {email.email}
                                </option>
                            )
                        }

                        return options
                    }, [] as React.ReactFragment[])}
                </select>{' '}
                <button type="submit" className="btn btn-primary" disabled={true}>
                    Save
                </button>
            </Form>
            {/* TODO: check this */}
            {status.errorDescription && (
                <ErrorAlert className="mt-2" error={status.errorDescription} history={history} />
            )}
        </div>
    )
}
