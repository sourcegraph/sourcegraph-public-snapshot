/* eslint-disable react/jsx-no-bind */
import React, { FunctionComponent, useEffect, useState, useCallback } from 'react'
import { RouteComponentProps } from 'react-router'
import * as H from 'history'

import { queryGraphQL } from '../../../backend/graphql'
import { UserAreaUserFields } from '../../../graphql-operations'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../../shared/src/util/errors'

import { eventLogger } from '../../../tracking/eventLogger'
import { ErrorAlert } from '../../../components/alerts'
import { PageTitle } from '../../../components/PageTitle'
import { UserEmail } from './UserEmail'
import { AddUserEmailForm } from './AddUserEmailForm'
import { SetUserPrimaryEmailForm } from './SetUserPrimaryEmailForm'

interface Props extends RouteComponentProps<{}> {
    user: UserAreaUserFields
    history: H.History
}

interface LoadingState {
    loading: boolean
    errorDescription: Error | null
}

export const UserEmailSettings: FunctionComponent<Props> = ({ user, history }) => {
    const [emails, setEmails] = useState<GQL.IUserEmail[]>([])
    const [status, setStatus] = useState<LoadingState>({ loading: false, errorDescription: null })

    const updateNewPrimaryEmail = (updatedEmail: string): GQL.IUserEmail[] =>
        emails.map(email => {
            if (email.isPrimary && email.email !== updatedEmail) {
                email.isPrimary = false
            }

            if (email.email === updatedEmail) {
                email.isPrimary = true
            }

            return email
        })

    const onEmailVerify = ({ email: verifiedEmail, verified }: { email: string; verified: boolean }): void => {
        const updatedEmails = emails.map(email => {
            if (email.email === verifiedEmail) {
                email.verified = verified
            }

            return email
        })

        setEmails(updatedEmails)
    }

    const onEmailRemove = (deletedEmail: string): void => {
        setEmails(emails.filter(({ email }) => email !== deletedEmail))
    }

    const onPrimaryEmailSet = (email: string): void => {
        setEmails(updateNewPrimaryEmail(email))
    }

    const fetchEmails = useCallback(async (): Promise<void> => {
        const { data, errors } = await queryGraphQL(
            gql`
                query UserEmails($user: ID!) {
                    node(id: $user) {
                        ... on User {
                            emails {
                                email
                                isPrimary
                                verified
                                verificationPending
                                viewerCanManuallyVerify
                            }
                        }
                    }
                }
            `,
            { user: user.id }
        ).toPromise()

        if (!data || (errors && errors.length > 0)) {
            setStatus({ loading: false, errorDescription: createAggregateError(errors) })
        } else {
            setStatus({ errorDescription: null, loading: false })
            const userResult = data.node as GQL.IUser
            setEmails(userResult.emails)
        }
    }, [user])

    useEffect(() => {
        eventLogger.logViewEvent('UserSettingsEmails')
    }, [])

    useEffect(() => {
        console.log('EmailSettings')
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        fetchEmails()
    }, [fetchEmails])

    return (
        <div className="user-settings-emails-page">
            <PageTitle title="Emails" />
            <h2>Emails</h2>
            {status.errorDescription && (
                <ErrorAlert className="mt-2" error={status.errorDescription} history={history} />
            )}
            <div className="list-group list-group-flush mt-3">
                {/* TODO: Fix this class */}
                <ul className="filtered-connection__nodes">
                    {emails.map(email => (
                        <li key={email.email} className="list-group-item py-2">
                            <UserEmail
                                user={user.id}
                                email={email}
                                onEmailVerify={onEmailVerify}
                                onDidRemove={onEmailRemove}
                                history={history}
                            />
                        </li>
                    ))}
                </ul>
            </div>
            <AddUserEmailForm className="mt-4" user={user.id} onDidAdd={fetchEmails} history={history} />
            <hr className="mt-4" />
            <SetUserPrimaryEmailForm
                className="mt-4"
                user={user.id}
                emails={emails}
                onDidSet={onPrimaryEmailSet}
                history={history}
            />
        </div>
    )
}
