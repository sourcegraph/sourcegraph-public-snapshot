import React, { useState, useCallback, FunctionComponent } from 'react'
import { Modal } from 'reactstrap'
import * as H from 'history'

import { mutateGraphQL } from '../../../backend/graphql'
import { IUserEmail } from '../../../../../shared/src/graphql/schema'
import { gql } from '../../../../../shared/src/graphql/graphql'

import { eventLogger } from '../../../tracking/eventLogger'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../components/alerts'
import { LoaderButton } from '../../../components/LoaderButton'

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
    const [modal, setModal] = useState(false)

    const toggleModal = useCallback(() => setModal(!modal), [modal])

    const removeEmail = useCallback(async (): Promise<void> => {
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

        setModal(false)
    }, [email, status, user, onDidRemove])

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

    let verifiedLinkFragment: React.ReactFragment
    if (verified) {
        verifiedLinkFragment = <span className="badge badge-success">Verified</span>
    } else if (verificationPending) {
        verifiedLinkFragment = <span className="badge badge-info">Verification pending</span>
    } else {
        verifiedLinkFragment = <span className="badge badge-secondary">Not verified</span>
    }

    const removeLinkFragment: React.ReactFragment = isPrimary ? (
        <a data-tooltip="Can't remove primary email" className="btn btn-link text-muted">
            Remove
        </a>
    ) : (
        <a className="btn btn-link text-danger" onClick={toggleModal}>
            Remove
        </a>
    )

    return (
        <>
            <div className="d-flex align-items-center justify-content-between">
                <div>
                    <strong>{email}</strong> &nbsp;{verifiedLinkFragment}&nbsp; &nbsp;
                    {isPrimary && <span className="badge badge-primary">Primary</span>}
                </div>
                <div>
                    {viewerCanManuallyVerify && (
                        <a className="btn btn-link text-primary" onClick={() => updateEmailVerification(!verified)}>
                            {verified ? 'Mark as unverified' : 'Mark as verified'}
                        </a>
                    )}
                    {removeLinkFragment}
                </div>
            </div>
            {status.errorDescription && (
                <ErrorAlert className="mt-2" error={status.errorDescription} history={history} />
            )}
            {modal && (
                <Modal
                    isOpen={modal}
                    toggle={toggleModal}
                    centered={true}
                    autoFocus={true}
                    keyboard={true}
                    fade={false}
                >
                    <div className="modal-header">
                        <h4 className="modal-title">Remove email</h4>
                        <button
                            type="button"
                            className="btn btn-icon"
                            data-dismiss="modal"
                            aria-label="Close"
                            onClick={toggleModal}
                        >
                            <span aria-hidden="true">&times;</span>
                        </button>
                    </div>
                    <div className="modal-body">
                        <p>
                            Remove the email address <strong>{email}</strong>?
                        </p>
                    </div>
                    <div className="modal-footer">
                        <LoaderButton
                            loading={status.loading}
                            onClick={removeEmail}
                            label="Remove"
                            type="button"
                            className="btn btn-danger"
                        />
                        <button type="button" className="btn btn-primary" onClick={toggleModal}>
                            Cancel
                        </button>
                    </div>
                </Modal>
            )}
        </>
    )
}
