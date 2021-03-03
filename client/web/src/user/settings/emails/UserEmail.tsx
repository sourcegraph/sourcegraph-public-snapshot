import React, { FunctionComponent } from 'react'
import {
    UserEmailsResult,
    RemoveUserEmailResult,
    RemoveUserEmailVariables,
    SetUserEmailVerifiedResult,
    SetUserEmailVerifiedVariables,
    ResendVerificationEmailResult,
    ResendVerificationEmailVariables,
} from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { gql, useMutation } from '@apollo/client'
import { FETCH_USER_EMAILS } from './UserSettingsEmailsPage'

const REMOVE_USER_EMAIL = gql`
    mutation RemoveUserEmail($user: ID!, $email: String!) {
        removeUserEmail(user: $user, email: $email) {
            alwaysNil
        }
    }
`

const SET_EMAIL_VERIFIED = gql`
    mutation SetUserEmailVerified($user: ID!, $email: String!, $verified: Boolean!) {
        setUserEmailVerified(user: $user, email: $email, verified: $verified) {
            alwaysNil
        }
    }
`

const RESEND_EMAIL_VERIFICATION = gql`
    mutation ResendVerificationEmail($user: ID!, $email: String!) {
        resendVerificationEmail(user: $user, email: $email) {
            alwaysNil
        }
    }
`

interface Props {
    user: string
    email: NonNullable<UserEmailsResult['node']>['emails'][number]
}

export const UserEmail: FunctionComponent<Props> = ({
    user,
    email: { email, isPrimary, verified, verificationPending, viewerCanManuallyVerify },
}) => {
    const refetchQueries = [{ query: FETCH_USER_EMAILS, variables: { user } }]
    const [verifyUserEmail, verifyEmailResponse] = useMutation<
        SetUserEmailVerifiedResult,
        SetUserEmailVerifiedVariables
    >(SET_EMAIL_VERIFIED, { refetchQueries })
    const [removeUserEmail, removeUserEmailResponse] = useMutation<RemoveUserEmailResult, RemoveUserEmailVariables>(
        REMOVE_USER_EMAIL,
        {
            onCompleted: () => eventLogger.log('UserEmailAddressDeleted'),
            refetchQueries,
        }
    )
    const [resendEmailVerification, resendEmailResponse] = useMutation<
        ResendVerificationEmailResult,
        ResendVerificationEmailVariables
    >(RESEND_EMAIL_VERIFICATION, {
        onCompleted: () => eventLogger.log('UserEmailAddressVerificationResent'),
        refetchQueries,
    })

    const isLoading = removeUserEmailResponse.loading || verifyEmailResponse.loading || resendEmailResponse.loading

    const updateEmailVerification = async (): Promise<void> => {
        await verifyUserEmail({ variables: { user, email, verified } })
        if (verified) {
            eventLogger.log('UserEmailAddressMarkedVerified')
        } else {
            eventLogger.log('UserEmailAddressMarkedUnverified')
        }
    }

    return (
        <>
            <div className="d-flex align-items-center justify-content-between">
                <div>
                    <span className="mr-2">{email}</span>
                    {verified && <span className="badge badge-success mr-1">Verified</span>}
                    {!verified && !verificationPending && (
                        <span className="badge badge-secondary mr-1">Not verified</span>
                    )}
                    {isPrimary && <span className="badge badge-primary mr-1">Primary</span>}
                    {!verified && verificationPending && (
                        <span>
                            <span className="user-settings-emails-page__dot">&bull;&nbsp;</span>
                            <button
                                type="button"
                                className="btn btn-link text-primary p-0"
                                onClick={() => resendEmailVerification({ variables: { user, email } })}
                                disabled={isLoading}
                            >
                                Resend verification email
                            </button>
                        </span>
                    )}
                </div>
                <div>
                    {!viewerCanManuallyVerify && (
                        <button
                            type="button"
                            className="btn btn-link text-primary p-0"
                            onClick={updateEmailVerification}
                            disabled={isLoading}
                        >
                            {verified ? 'Mark as unverified' : 'Mark as verified'}
                        </button>
                    )}{' '}
                    {!isPrimary && (
                        <button
                            type="button"
                            className="btn btn-link text-danger p-0"
                            onClick={() => removeUserEmail({ variables: { user, email } })}
                            disabled={isLoading}
                        >
                            Remove
                        </button>
                    )}
                </div>
            </div>
        </>
    )
}
