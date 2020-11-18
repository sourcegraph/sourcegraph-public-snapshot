/* eslint-disable react/jsx-no-bind */
import React, { FunctionComponent, useEffect, useState, useCallback } from 'react'
import { RouteComponentProps } from 'react-router'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import { Subscription } from 'rxjs'

import { queryGraphQL } from '../../../backend/graphql'
import { UserAreaUserFields } from '../../../graphql-operations'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { SiteFlags } from '../../../site'
import { siteFlags } from '../../../site/backend'

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

export const UserSettingsEmailsPage: FunctionComponent<Props> = ({ user, history }) => {
    const [emails, setEmails] = useState<GQL.IUserEmail[]>([])
    const [status, setStatus] = useState<LoadingState>({ loading: false, errorDescription: null })
    const [flags, setFlags] = useState<SiteFlags>()

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
        setStatus({ errorDescription: null, loading: true })

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
    }, [user, setStatus, setEmails])

    useEffect(() => {
        eventLogger.logViewEvent('UserSettingsEmails')
        const subscriptions = new Subscription()
        subscriptions.add(siteFlags.subscribe(setFlags))

        return () => {
            subscriptions.unsubscribe()
        }
    }, [])

    useEffect(() => {
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        fetchEmails()
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

            {status.errorDescription && (
                <ErrorAlert className="mt-2" error={status.errorDescription} history={history} />
            )}

            {status.loading ? (
                <span className="filtered-connection__loader">
                    <LoadingSpinner className="icon-inline" />
                </span>
            ) : (
                <div className="list-group list-group-flush mt-3">
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
            )}

            {/* re-fetch emails when new emails are added to guarantee correct state */}
            <AddUserEmailForm className="mt-4" user={user.id} onDidAdd={fetchEmails} history={history} />
            <hr className="mt-4" />
            {!status.loading && (
                <SetUserPrimaryEmailForm
                    className="mt-4"
                    user={user.id}
                    emails={emails}
                    onDidSet={onPrimaryEmailSet}
                    history={history}
                />
            )}
        </div>
    )
}
