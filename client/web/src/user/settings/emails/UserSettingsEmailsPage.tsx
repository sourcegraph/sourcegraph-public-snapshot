import { FunctionComponent, useEffect, useState, useCallback } from 'react'

import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'
import { Container, PageHeader, LoadingSpinner, useObservable, Alert } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../../backend/graphql'
import { PageTitle } from '../../../components/PageTitle'
import { Scalars, UserEmailsResult, UserEmailsVariables, UserSettingsAreaUserFields } from '../../../graphql-operations'
import { siteFlags } from '../../../site/backend'
import { eventLogger } from '../../../tracking/eventLogger'

import { AddUserEmailForm } from './AddUserEmailForm'
import { SetUserPrimaryEmailForm } from './SetUserPrimaryEmailForm'
import { UserEmail } from './UserEmail'

import styles from './UserSettingsEmailsPage.module.scss'

interface Props {
    user: UserSettingsAreaUserFields
}

type UserEmail = (NonNullable<UserEmailsResult['node']> & { __typename: 'User' })['emails'][number]
type Status = undefined | 'loading' | 'loaded' | ErrorLike
type EmailActionError = undefined | ErrorLike

export const UserSettingsEmailsPage: FunctionComponent<React.PropsWithChildren<Props>> = ({ user }) => {
    const [emails, setEmails] = useState<UserEmail[]>([])
    const [statusOrError, setStatusOrError] = useState<Status>()
    const [emailActionError, setEmailActionError] = useState<EmailActionError>()

    const onEmailRemove = useCallback(
        (deletedEmail: string): void => {
            setEmails(emails => emails.filter(({ email }) => email !== deletedEmail))
            // always cleanup email action errors when removing emails
            setEmailActionError(undefined)
        },
        [setEmailActionError]
    )

    const fetchEmails = useCallback(async (): Promise<void> => {
        setStatusOrError('loading')

        const fetchedEmails = await fetchUserEmails(user.id)

        // always cleanup email action errors when re-fetching emails
        setEmailActionError(undefined)

        if (fetchedEmails?.node?.__typename === 'User' && fetchedEmails.node.emails) {
            setEmails(fetchedEmails.node.emails)
            setStatusOrError('loaded')
        } else {
            setStatusOrError(asError("Sorry, we couldn't fetch user emails. Try again?"))
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

    if (statusOrError === 'loading') {
        return <LoadingSpinner />
    }

    return (
        <div className={styles.userSettingsEmailsPage} data-testid="user-settings-emails-page">
            <PageTitle title="Emails" />
            <PageHeader headingElement="h2" path={[{ text: 'Emails' }]} className="mb-3" />

            {flags && !flags.sendsEmailVerificationEmails && (
                <Alert variant="warning">
                    Sourcegraph is not configured to send email verifications. Newly added email addresses must be
                    manually verified by a site admin.
                </Alert>
            )}

            {isErrorLike(statusOrError) && <ErrorAlert className="mt-2" error={statusOrError} />}
            {isErrorLike(emailActionError) && <ErrorAlert className="mt-2" error={emailActionError} />}

            <Container>
                <ul className="list-group">
                    {emails.map(email => (
                        <li key={email.email} className={classNames('list-group-item', styles.listItem)}>
                            <UserEmail
                                user={user.id}
                                email={email}
                                onEmailVerify={fetchEmails}
                                onEmailResendVerification={fetchEmails}
                                onDidRemove={onEmailRemove}
                                onError={setEmailActionError}
                            />
                        </li>
                    ))}
                    {emails.length === 0 && (
                        <li className={classNames('list-group-item text-muted', styles.listItem)}>No emails</li>
                    )}
                </ul>
            </Container>
            {/* re-fetch emails on onDidAdd to guarantee correct state */}
            <AddUserEmailForm className={styles.emailForm} user={user.id} onDidAdd={fetchEmails} />
            <hr className="my-4" />
            <SetUserPrimaryEmailForm user={user.id} emails={emails} onDidSet={fetchEmails} />
        </div>
    )
}

async function fetchUserEmails(userID: Scalars['ID']): Promise<UserEmailsResult> {
    return dataOrThrowErrors(
        await requestGraphQL<UserEmailsResult, UserEmailsVariables>(
            gql`
                query UserEmails($user: ID!) {
                    node(id: $user) {
                        ... on User {
                            __typename
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
            { user: userID }
        ).toPromise()
    )
}
