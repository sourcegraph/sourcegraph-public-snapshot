/* eslint-disable react/jsx-no-bind */
import React, { FunctionComponent, useEffect, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import * as H from 'history'

import { queryGraphQL } from '../../../backend/graphql'
import { UserAreaUserFields } from '../../../graphql-operations'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../../shared/src/util/errors'

import { PageTitle } from '../../../components/PageTitle'
import { UserEmail } from './UserEmail'
import { AddUserEmailForm } from './AddUserEmailForm'
import { SetUserPrimaryEmailForm } from './SetUserPrimaryEmailForm'

interface Props extends RouteComponentProps<{}> {
    user: UserAreaUserFields
    history: H.History
}

export const UserEmailSettings: FunctionComponent<Props> = ({ user, history }) => {
    const [emails, setEmails] = useState<GQL.IUserEmail[]>([])

    const onEmailAdd = (email: string): void => {
        const newEmail = {
            email,
            isPrimary: false,
            verified: false,
            verificationPending: true,
        }

        // TODO: pick types instead of IUserEmail
        setEmails([...emails, newEmail])
    }

    const onEmailRemove = (deletedEmail: string): void => {
        // seems like page needs tp be re-fetched or maybe use optimistc?
        setEmails(emails.filter(({ email }) => email !== deletedEmail))
    }

    const onPrimaryEmailSet = (email: string): void => {
        console.log(email)
        // find current primary email
        // set as non-primary
        // find the one we just set and make it a primary
        // setEmails
    }

    useEffect(() => {
        const fetchUserEmails = async (): Promise<void> => {
            // TODO
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

            if (errors || !data || !data?.node) {
                // TODO: why?
                throw createAggregateError(errors)
            }

            const userResult = data.node as GQL.IUser

            setEmails(userResult.emails)
        }

        // TODO
        fetchUserEmails().catch(error => {
            throw error
        })
    }, [user.id, setEmails])

    return (
        <div className="user-settings-emails-page">
            <PageTitle title="Emails" />
            <h2>Emails</h2>
            <div className="list-group list-group-flush mt-3">
                {/* TODO: Fix this class */}
                <ul className="filtered-connection__nodes">
                    {emails.map(email => (
                        // TODO: fix this
                        <li key={email.email} className="list-group-item py-2">
                            <UserEmail user={user.id} email={email} onDidRemove={onEmailRemove} history={history} />
                        </li>
                    ))}
                </ul>
            </div>
            <AddUserEmailForm className="mt-4" user={user.id} onDidAdd={onEmailAdd} history={history} />
            <hr className="mt-4" />
            <SetUserPrimaryEmailForm
                className="mt-4"
                userId={user.id}
                emails={emails}
                onDidSet={onPrimaryEmailSet}
                history={history}
            />
        </div>
    )
}
