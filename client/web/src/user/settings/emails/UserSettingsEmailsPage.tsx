import React, { FunctionComponent, useEffect } from 'react'
import { RouteComponentProps } from 'react-router'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import { gql, useQuery } from '@apollo/client'

import { UserAreaUserFields, UserEmailsResult, UserEmailsVariables } from '../../../graphql-operations'
import { useObservable } from '../../../../../shared/src/util/useObservable'
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

type UserEmail = NonNullable<UserEmailsResult['node']>['emails'][number]

export const FETCH_USER_EMAILS = gql`
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
`

export const UserSettingsEmailsPage: FunctionComponent<Props> = ({ user, history }) => {
    const { data, loading, error } = useQuery<UserEmailsResult, UserEmailsVariables>(FETCH_USER_EMAILS, {
        variables: { user: user.id },
    })
    const flags = useObservable(siteFlags)

    useEffect(() => {
        eventLogger.logViewEvent('UserSettingsEmails')
    }, [])

    return (
        <div className="user-settings-emails-page">
            <PageTitle title="Emails" />

            {flags && !flags.sendsEmailVerificationEmails && (
                <div className="alert alert-warning mt-2">
                    Sourcegraph is not configured to send email verifications. Newly added email addresses must be
                    manually verified by a site admin.
                </div>
            )}

            {error && <ErrorAlert className="mt-2" error={error} history={history} />}

            <h2>Emails</h2>

            {loading ? (
                <div className="d-flex justify-content-center">
                    <LoadingSpinner className="icon-inline" />
                </div>
            ) : (
                <div className="mt-4">
                    <ul className="list-group">
                        {data?.node?.emails.map(email => (
                            <li key={email.email} className="list-group-item p-3">
                                <UserEmail user={user.id} email={email} />
                            </li>
                        ))}
                    </ul>
                </div>
            )}

            <AddUserEmailForm className="mt-4" user={user.id} history={history} />
            <hr className="my-4" />
            {data?.node && <SetUserPrimaryEmailForm user={user.id} emails={data.node.emails} history={history} />}
        </div>
    )
}
