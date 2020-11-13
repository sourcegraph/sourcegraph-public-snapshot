import React, { useState, FunctionComponent } from 'react'
import classNames from 'classnames'
import * as H from 'history'

import { mutateGraphQL } from '../../../backend/graphql'
import { IUserEmail } from '../../../../../shared/src/graphql/schema'
import { gql } from '../../../../../shared/src/graphql/graphql'

import { eventLogger } from '../../../tracking/eventLogger'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../components/alerts'

interface Props {
    user: string
    email: IUserEmail
    history: H.History
    onDidRemove?: (email: string) => void
}

interface UserEmailState {
    loading: boolean
    errorDescription: Error | null
}

export const UserEmail: FunctionComponent<Props> = ({
    user,
    email: { email, isPrimary, verified, verificationPending },
    onDidRemove,
    history,
}) => {
    const [status, setStatus] = useState<UserEmailState>({ loading: false, errorDescription: null })

    let verifiedFragment: React.ReactFragment
    if (verified) {
        verifiedFragment = <span className="badge badge-success">Verified</span>
    } else if (verificationPending) {
        verifiedFragment = <span className="badge badge-info">Verification pending</span>
    } else {
        verifiedFragment = <span className="badge badge-secondary">Not verified</span>
    }

    const removeUserEmail = async (): Promise<void> => {
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

        // TODO: check this
        if (!data || (errors && errors.length > 0)) {
            const aggregateError = createAggregateError(errors)
            setStatus({ loading: false, errorDescription: aggregateError })
            throw aggregateError
        }

        setStatus({ ...status, loading: false })

        if (onDidRemove) {
            onDidRemove(email)
        }
        eventLogger.log('UserEmailAddressDeleted')
    }

    const removeLinkClasses = classNames({
        btn: true,
        'text-danger': !status.loading,
        'text-muted': status.loading,
    })

    return (
        <div className="d-flex align-items-center justify-content-between">
            <div>
                <strong>{email}</strong> &nbsp;{verifiedFragment}&nbsp; &nbsp;
                {isPrimary && <span className="badge badge-primary">Primary</span>}
            </div>
            {verified && (
                <a className={removeLinkClasses} onClick={status.loading ? () => {} : removeUserEmail}>
                    Remove
                </a>
            )}
            {/* TODO: check this */}
            {status.errorDescription && (
                <ErrorAlert className="mt-2" error={status.errorDescription} history={history} />
            )}
        </div>
    )
}
