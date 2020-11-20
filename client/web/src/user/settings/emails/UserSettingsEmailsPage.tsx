import React, { FunctionComponent, useEffect, useState, useCallback } from 'react'
import { RouteComponentProps } from 'react-router'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'

import { requestGraphQL } from '../../../backend/graphql'
import { UserAreaUserFields, UserEmailsResult, UserEmailsVariables } from '../../../graphql-operations'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { siteFlags } from '../../../site/backend'
import { asError } from '../../../../../shared/src/util/errors'

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
    error?: Error
}

type UserEmail = Omit<GQL.IUserEmail, '__typename' | 'user'>

export const UserSettingsEmailsPage: FunctionComponent<Props> = ({ user, history }) => {
    const [emails, setEmails] = useState<UserEmail[]>([])
    const [status, setStatus] = useState<LoadingState>({ loading: false })

    const onEmailVerify = useCallback(
        ({ email: verifiedEmail, verified }: { email: string; verified: boolean }): void => {
            const updatedEmails = emails.map(email => {
                if (email.email === verifiedEmail) {
                    email.verified = verified
                }

                return email
            })

            setEmails(updatedEmails)
        },
        [emails]
    )

    const onEmailRemove = useCallback(
        (deletedEmail: string): void => {
            setEmails(emails.filter(({ email }) => email !== deletedEmail))
        },
        [emails]
    )

    const onPrimaryEmailSet = useCallback(
        (email: string): void => {
            const updateNewPrimaryEmail = (updatedEmail: string): UserEmail[] =>
                emails.map(email => {
                    if (email.isPrimary && email.email !== updatedEmail) {
                        email.isPrimary = false
                    }

                    if (email.email === updatedEmail) {
                        email.isPrimary = true
                    }

                    return email
                })

            setEmails(updateNewPrimaryEmail(email))
        },
        [emails]
    )

    const fetchEmails = useCallback(async (): Promise<void> => {
        setStatus({ loading: true })
        let fetchedEmails

        try {
            fetchedEmails = dataOrThrowErrors(
                await requestGraphQL<UserEmailsResult, UserEmailsVariables>(
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
            )
        } catch (error) {
            setStatus({ loading: false, error: asError(error) })
        }

        // TODO: check this logic
        if (fetchedEmails?.node?.emails) {
            setStatus({ loading: false })
            setEmails(fetchedEmails.node.emails)
        }
    }, [user, setStatus, setEmails])

    const flags = useObservable(siteFlags)

    useEffect(() => {
        eventLogger.logViewEvent('UserSettingsEmails')
    }, [])

    useEffect(() => {
        fetchEmails().catch(error => {
            setStatus({ loading: false, error: asError(error) })
        })
    }, [fetchEmails])

    return (
        <div className="user-settings-emails-page">
            <PageTitle title="Emails" />
            <h2>Emails</h2>

            {flags && !flags.sendsEmailVerificationEmails && (
                <div className="alert alert-warning mt-2">
                    Sourcegraph is not configured to send email verifications. Newly added email addresses must be
                    manually verified by a site admin.
                </div>
            )}

            {status.error && <ErrorAlert className="mt-2" error={status.error} history={history} />}

            {status.loading ? (
                <span className="filtered-connection__loader">
                    <LoadingSpinner className="icon-inline" />
                </span>
            ) : (
                <div className="list-group list-group-flush mt-4">
                    <ul className="user-settings-emails-page__list">
                        {emails.map(email => (
                            <li key={email.email} className="list-group-item p-3">
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
            )}

            {/* re-fetch emails when new emails are added to guarantee correct state */}
            <AddUserEmailForm className="mt-4" user={user.id} onDidAdd={fetchEmails} history={history} />
            <hr className="my-4" />
            {!status.loading && (
                <SetUserPrimaryEmailForm
                    user={user.id}
                    emails={emails}
                    onDidSet={onPrimaryEmailSet}
                    history={history}
                />
            )}
        </div>
    )
}
