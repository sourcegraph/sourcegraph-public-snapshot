import React, { FunctionComponent, useEffect, useState, useCallback } from 'react'
import { RouteComponentProps } from 'react-router'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'

import { requestGraphQL } from '../../../backend/graphql'
import { UserAreaUserFields, UserEmailsResult, UserEmailsVariables } from '../../../graphql-operations'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { siteFlags } from '../../../site/backend'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
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

type UserEmail = NonNullable<UserEmailsResult['node']>['emails'][number]
type Status = undefined | 'loading' | 'loaded' | ErrorLike

export const UserSettingsEmailsPage: FunctionComponent<Props> = ({ user, history }) => {
    const [emails, setEmails] = useState<UserEmail[]>([])
    const [statusOrError, setStatusOrError] = useState<Status>()

    const onEmailRemove = useCallback(
        (deletedEmail: string): void => {
            setEmails(emails.filter(({ email }) => email !== deletedEmail))
        },
        [emails]
    )

    const fetchEmails = useCallback(async (): Promise<void> => {
        setStatusOrError('loading')

        try {
            const fetchedEmails = dataOrThrowErrors(
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

            if (fetchedEmails?.node?.emails) {
                setEmails(fetchedEmails.node.emails)
                setStatusOrError('loaded')
            }
        } catch (error) {
            setStatusOrError(asError(error))
        }
    }, [user, setStatusOrError, setEmails])

    const flags = useObservable(siteFlags)

    useEffect(() => {
        eventLogger.logViewEvent('UserSettingsEmails')
    }, [])

    useEffect(() => {
        fetchEmails().catch(error => {
            setStatusOrError(asError(error))
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

            {isErrorLike(statusOrError) && <ErrorAlert className="mt-2" error={statusOrError} history={history} />}

            {statusOrError === 'loading' ? (
                <span className="user-settings-emails-page__loader">
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
                                    onEmailVerify={fetchEmails}
                                    onDidRemove={onEmailRemove}
                                    history={history}
                                />
                            </li>
                        ))}
                    </ul>
                </div>
            )}

            <AddUserEmailForm
                key={emails.length}
                className="mt-4"
                user={user.id}
                onDidAdd={fetchEmails}
                history={history}
            />
            <hr className="my-4" />
            {statusOrError === 'loaded' && (
                <SetUserPrimaryEmailForm user={user.id} emails={emails} onDidSet={fetchEmails} history={history} />
            )}
        </div>
    )
}
