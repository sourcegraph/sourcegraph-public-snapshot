import React, { useState, FunctionComponent } from 'react'
import classNames from 'classnames'
import * as H from 'history'

import { mutateGraphQL } from '../../../backend/graphql'
import { IUserEmail } from '../../../../../shared/src/graphql/schema'
import { gql } from '../../../../../shared/src/graphql/graphql'

import { eventLogger } from '../../../tracking/eventLogger'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../components/alerts'

interface VerificationUpdate {
    email: string
    verified: boolean
}

interface Props {
    user: string
    email: IUserEmail
    history: H.History

    onDidRemove?: (email: string) => void
    onEmailVerify?: (update: VerificationUpdate) => void
}

interface LoadingState {
    loading: boolean
    errorDescription: Error | null
}

export const UserEmail: FunctionComponent<Props> = ({
    user,
    email: { email, isPrimary, verified, verificationPending, viewerCanManuallyVerify },
    onDidRemove,
    onEmailVerify,
    history,
}) => {
    const [status, setStatus] = useState<LoadingState>({ loading: false, errorDescription: null })

    let verifiedFragment: React.ReactFragment
    if (verified) {
        verifiedFragment = <span className="badge badge-success">Verified</span>
    } else if (verificationPending) {
        verifiedFragment = <span className="badge badge-info">Verification pending</span>
    } else {
        verifiedFragment = <span className="badge badge-secondary">Not verified</span>
    }

    const removeEmail = async (): Promise<void> => {
        if (!window.confirm(`Remove the email address ${email}?`)) {
            return
        }

        setStatus({ ...status, loading: true })

        const { data, errors } = await mutateGraphQL(
            gql`
                mutation RemoveUserEmail($user: ID!, $email: String!) {
                    removeUserEmail(user: $user, email: $email) {
                        alwaysNil
                    }
                }
            `,
            { user, email }
        ).toPromise()

        if (!data || (errors && errors.length > 0)) {
            setStatus({ loading: false, errorDescription: createAggregateError(errors) })
        } else {
            setStatus({ ...status, loading: false })
            eventLogger.log('UserEmailAddressDeleted')

            if (onDidRemove) {
                onDidRemove(email)
            }
        }
    }

    const updateEmailVerification = async (verified: boolean): Promise<void> => {
        setStatus({ ...status, loading: true })

        const { data, errors } = await mutateGraphQL(
            gql`
                mutation SetUserEmailVerified($user: ID!, $email: String!, $verified: Boolean!) {
                    setUserEmailVerified(user: $user, email: $email, verified: $verified) {
                        alwaysNil
                    }
                }
            `,
            { user, email, verified }
        ).toPromise()

        if (!data || (errors && errors.length > 0)) {
            setStatus({ loading: false, errorDescription: createAggregateError(errors) })
        } else {
            setStatus({ ...status, loading: false })

            if (verified) {
                eventLogger.log('UserEmailAddressMarkedVerified')
            } else {
                eventLogger.log('UserEmailAddressMarkedUnverified')
            }

            if (onEmailVerify) {
                onEmailVerify({ email, verified })
            }
        }
    }

    return (
        <>
            <div className="d-flex align-items-center justify-content-between">
                <div>
                    <strong>{email}</strong> &nbsp;{verifiedFragment}&nbsp; &nbsp;
                    {isPrimary && <span className="badge badge-primary">Primary</span>}
                </div>
                <div>
                    {viewerCanManuallyVerify ? (
                        <a className="btn btn-link text-primary" onClick={() => updateEmailVerification(!verified)}>
                            {verified ? 'Mark as unverified' : 'Mark as verified'}
                        </a>
                    ) : null}
                    {!isPrimary ? (
                        <a className="btn btn-link text-danger" onClick={!status.loading ? removeEmail : () => {}}>
                            Remove
                        </a>
                    ) : null}
                </div>
            </div>
            {status.errorDescription && (
                <ErrorAlert className="mt-2" error={status.errorDescription} history={history} />
            )}
        </>
    )
}
