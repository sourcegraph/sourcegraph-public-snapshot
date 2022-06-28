import { useState, FunctionComponent } from 'react'

import { asError, ErrorLike } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { Badge, Button, screenReaderAnnounce } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../../backend/graphql'
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

import styles from './UserEmail.module.scss'

interface Props {
    user: string
    email: (NonNullable<UserEmailsResult['node']> & { __typename: 'User' })['emails'][number]
    onError: (error: ErrorLike) => void

    onDidRemove?: (email: string) => void
    onEmailVerify?: () => void
    onEmailResendVerification?: () => void
}

export const UserEmail: FunctionComponent<React.PropsWithChildren<Props>> = ({
    user,
    email: { email, isPrimary, verified, verificationPending, viewerCanManuallyVerify },
    onError,
    onDidRemove,
    onEmailVerify,
    onEmailResendVerification,
}) => {
    const [isLoading, setIsLoading] = useState(false)

    const handleError = (error: ErrorLike): void => {
        onError(asError(error))
        setIsLoading(false)
    }

    const removeEmail = async (): Promise<void> => {
        setIsLoading(true)

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

            setIsLoading(false)
            eventLogger.log('UserEmailAddressDeleted')
            screenReaderAnnounce('Email address removed')

            if (onDidRemove) {
                onDidRemove(email)
            }
        } catch (error) {
            handleError(error)
        }
    }

    const updateEmailVerification = async (verified: boolean): Promise<void> => {
        setIsLoading(true)

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

            setIsLoading(false)

            if (verified) {
                eventLogger.log('UserEmailAddressMarkedVerified')
            } else {
                eventLogger.log('UserEmailAddressMarkedUnverified')
            }

            if (onEmailVerify) {
                onEmailVerify()
            }
        } catch (error) {
            handleError(error)
        }
    }

    const resendEmailVerification = async (email: string): Promise<void> => {
        setIsLoading(true)

        try {
            dataOrThrowErrors(
                await requestGraphQL<ResendVerificationEmailResult, ResendVerificationEmailVariables>(
                    gql`
                        mutation ResendVerificationEmail($user: ID!, $email: String!) {
                            resendVerificationEmail(user: $user, email: $email) {
                                alwaysNil
                            }
                        }
                    `,
                    { user, email }
                ).toPromise()
            )

            setIsLoading(false)
            eventLogger.log('UserEmailAddressVerificationResent')

            onEmailResendVerification?.()
        } catch (error) {
            handleError(error)
        }
    }

    return (
        <>
            <div className="d-flex align-items-center justify-content-between">
                <div className="d-flex align-items-center">
                    <span className="mr-2">{email}</span>
                    {/*
                        a11y-ignore
                        Rule: "color-contrast" (Elements must have sufficient color contrast)
                        GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
                    */}
                    {verified && (
                        <Badge variant="success" className="mr-1 a11y-ignore">
                            Verified
                        </Badge>
                    )}
                    {!verified && !verificationPending && (
                        <Badge variant="secondary" className="mr-1">
                            Not verified
                        </Badge>
                    )}
                    {isPrimary && (
                        <Badge variant="primary" className="mr-1">
                            Primary
                        </Badge>
                    )}
                    {!verified && verificationPending && (
                        <span>
                            <span className={styles.dot}>&bull;&nbsp;</span>
                            <Button
                                className="p-0"
                                onClick={() => resendEmailVerification(email)}
                                disabled={isLoading}
                                variant="link"
                            >
                                Resend verification email
                            </Button>
                        </span>
                    )}
                </div>
                <div className="d-flex align-items-center">
                    {viewerCanManuallyVerify && (
                        <Button
                            className="p-0"
                            onClick={() => updateEmailVerification(!verified)}
                            disabled={isLoading}
                            variant="link"
                        >
                            {verified ? 'Mark as unverified' : 'Mark as verified'}
                        </Button>
                    )}{' '}
                    {!isPrimary && (
                        <Button className="text-danger p-0" onClick={removeEmail} disabled={isLoading} variant="link">
                            Remove
                        </Button>
                    )}
                </div>
            </div>
        </>
    )
}
