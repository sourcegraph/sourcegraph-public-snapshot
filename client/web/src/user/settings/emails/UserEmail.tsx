import React, { useState, useCallback, FunctionComponent } from 'react'
import * as H from 'history'

import { requestGraphQL } from '../../../backend/graphql'
import { IUserEmail } from '../../../../../shared/src/graphql/schema'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { asError } from '../../../../../shared/src/util/errors'

import { eventLogger } from '../../../tracking/eventLogger'
import { ErrorAlert } from '../../../components/alerts'

import {
    RemoveUserEmailResult,
    RemoveUserEmailVariables,
    SetUserEmailVerifiedResult,
    SetUserEmailVerifiedVariables,
} from '../../../graphql-operations'

interface VerificationUpdate {
    email: string
    verified: boolean
}

interface Props {
    user: string
    email: Omit<IUserEmail, '__typename' | 'user'>
    history: H.History

    onDidRemove?: (email: string) => void
    onEmailVerify?: (update: VerificationUpdate) => void
}

interface LoadingState {
    loading: boolean
    error?: Error
}

export const UserEmail: FunctionComponent<Props> = ({
    user,
    email: { email, isPrimary, verified, verificationPending, viewerCanManuallyVerify },
    onDidRemove,
    onEmailVerify,
    history,
}) => {
    const [status, setStatus] = useState<LoadingState>({ loading: false })

    const removeEmail = useCallback(async (): Promise<void> => {
        setStatus({ loading: true })

        try {
            dataOrThrowErrors(
                await requestGraphQL<RemoveUserEmailResult, RemoveUserEmailVariables>(
                    gql`
                        mutation RemoveUserEmail($user: ID!, $email: String!) {
                            removeUserEmail(user: $user, email: $email) {
                                alwaysNil
                            }
                        }
                    `,
                    { user, email }
                ).toPromise()
            )
        } catch (error) {
            setStatus({ loading: false, error: asError(error) })
            return
        }

        setStatus({ loading: false })
        eventLogger.log('UserEmailAddressDeleted')

        if (onDidRemove) {
            onDidRemove(email)
        }
    }, [email, user, onDidRemove])

    const updateEmailVerification = async (verified: boolean): Promise<void> => {
        setStatus({ ...status, loading: true })

        try {
            dataOrThrowErrors(
                await requestGraphQL<SetUserEmailVerifiedResult, SetUserEmailVerifiedVariables>(
                    gql`
                        mutation SetUserEmailVerified($user: ID!, $email: String!, $verified: Boolean!) {
                            setUserEmailVerified(user: $user, email: $email, verified: $verified) {
                                alwaysNil
                            }
                        }
                    `,
                    { user, email, verified }
                ).toPromise()
            )
        } catch (error) {
            setStatus({ loading: false, error: asError(error) })
            return
        }

        setStatus({ loading: false })

        if (verified) {
            eventLogger.log('UserEmailAddressMarkedVerified')
        } else {
            eventLogger.log('UserEmailAddressMarkedUnverified')
        }

        if (onEmailVerify) {
            onEmailVerify({ email, verified })
        }
    }

    let verifiedLinkFragment: React.ReactFragment
    if (verified) {
        verifiedLinkFragment = <span className="badge badge-success">Verified</span>
    } else if (verificationPending) {
        verifiedLinkFragment = <span className="badge badge-info">Verification pending</span>
    } else {
        verifiedLinkFragment = <span className="badge badge-secondary">Not verified</span>
    }

    return (
        <>
            <div className="d-flex align-items-center justify-content-between">
                <div>
                    <span className="user-settings-emails-page__email">{email}</span> {verifiedLinkFragment}{' '}
                    {isPrimary && <span className="badge badge-primary">Primary</span>}
                </div>
                <div>
                    {viewerCanManuallyVerify && (
                        <button
                            type="button"
                            className="btn btn-link text-primary user-settings-emails-page__btn"
                            onClick={() => updateEmailVerification(!verified)}
                            disabled={status.loading}
                        >
                            {verified ? 'Mark as unverified' : 'Mark as verified'}
                        </button>
                    )}
                    {!isPrimary && (
                        <button
                            type="button"
                            className="btn btn-link text-danger user-settings-emails-page__btn"
                            onClick={removeEmail}
                            disabled={status.loading}
                        >
                            Remove
                        </button>
                    )}
                </div>
            </div>
            {status.error && <ErrorAlert className="mt-2" error={status.error} history={history} />}
        </>
    )
}
