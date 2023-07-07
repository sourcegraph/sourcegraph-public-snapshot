import { useState, useCallback, useEffect } from 'react'

import { mdiEmail } from '@mdi/js'
import classNames from 'classnames'
import { Navigate } from 'react-router-dom'

import { asError, ErrorLike } from '@sourcegraph/common'
import { gql, useMutation } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Checkbox, Form, H3, Modal, Text, Button, Icon } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { CodyColorIcon } from '../../cody/chat/CodyPageIcon'
import { isEmailVerificationNeededForCody } from '../../cody/isCodyEnabled'
import { LoaderButton } from '../../components/LoaderButton'
import {
    SubmitCodySurveyResult,
    SubmitCodySurveyVariables,
    setCompletedPostSignupVariables,
    setCompletedPostSignupResult,
} from '../../graphql-operations'
import { PageRoutes } from '../../routes.constants'
import { resendVerificationEmail } from '../../user/settings/emails/UserEmail'

import styles from './CodySurveyToast.module.scss'

const SUBMIT_CODY_SURVEY = gql`
    mutation SubmitCodySurvey($isForWork: Boolean!, $isForPersonal: Boolean!) {
        submitCodySurvey(isForWork: $isForWork, isForPersonal: $isForPersonal) {
            alwaysNil
        }
    }
`

const SET_COMPLETED_POST_SIGNUP = gql`
    mutation setCompletedPostSignup($userID: ID!) {
        setCompletedPostSignup(userID: $userID) {
            alwaysNil
        }
    }
`

const CodySurveyToastInner: React.FC<{ onSubmitEnd: () => void } & TelemetryProps> = ({
    onSubmitEnd,
    telemetryService,
}) => {
    const [isCodyForWork, setIsCodyForWork] = useState(false)
    const [isCodyForPersonalStuff, setIsCodyForPersonalStuff] = useState(false)

    const handleCodyForWorkChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setIsCodyForWork(event.target.checked)
    }, [])
    const handleCodyForPersonalStuffChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setIsCodyForPersonalStuff(event.target.checked)
    }, [])

    const [submitCodySurvey, { loading }] = useMutation<SubmitCodySurveyResult, SubmitCodySurveyVariables>(
        SUBMIT_CODY_SURVEY,
        {
            variables: {
                isForWork: isCodyForWork,
                isForPersonal: isCodyForPersonalStuff,
            },
        }
    )

    const handleSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>) => {
            const eventParams = { isCodyForPersonalStuff, isCodyForWork }
            telemetryService.log('CodyUsageToastSubmitted', eventParams, eventParams)
            event.preventDefault()
            // eslint-disable-next-line no-console
            submitCodySurvey().catch(console.error).finally(onSubmitEnd)
        },
        [isCodyForPersonalStuff, isCodyForWork, onSubmitEnd, submitCodySurvey, telemetryService]
    )

    useEffect(() => {
        telemetryService.log('CodySurveyToastViewed')
    }, [telemetryService])

    return (
        <Modal className={styles.codySurveyToastModal} position="center" aria-label="Welcome message">
            <H3 className="mb-4 d-flex align-items-center">
                <CodyColorIcon className={styles.codyIcon} />
                <span>Just one more thing...</span>
            </H3>
            <Text className="mb-3">How will you be using Cody, our AI assistant?</Text>
            <Form onSubmit={handleSubmit}>
                <Checkbox
                    id="cody-for-work"
                    label="for work"
                    wrapperClassName="mb-2"
                    checked={isCodyForWork}
                    disabled={loading}
                    onChange={handleCodyForWorkChange}
                    className={styles.modalCheckbox}
                />
                <Checkbox
                    id="cody-for-personal"
                    label="for personal stuff"
                    wrapperClassName="mb-2"
                    checked={isCodyForPersonalStuff}
                    disabled={loading}
                    onChange={handleCodyForPersonalStuffChange}
                    className={styles.modalCheckbox}
                />
                <div className="d-flex justify-content-end">
                    <LoaderButton
                        className={styles.codySurveyToastModalButton}
                        type="submit"
                        loading={loading}
                        label="Get started"
                        disabled={!(isCodyForPersonalStuff || isCodyForWork)}
                    />
                </div>
            </Form>
        </Modal>
    )
}

const CodyVerifyEmailToast: React.FC<
    { onNext: () => void; authenticatedUser: AuthenticatedUser; hasVerifiedEmail: boolean } & TelemetryProps
> = ({ onNext, authenticatedUser, telemetryService, hasVerifiedEmail }) => {
    const [sending, setSending] = useState(false)
    const [resentEmailTo, setResentEmailTo] = useState<string | null>(null)
    const [resendEmailError, setResendEmailError] = useState<ErrorLike | null>(null)
    const resend = useCallback(async () => {
        const email = (authenticatedUser.emails || []).find(({ verified }) => !verified)?.email
        if (email) {
            setSending(true)
            await resendVerificationEmail(authenticatedUser.id, email, {
                onSuccess: () => {
                    setResentEmailTo(email)
                    setResendEmailError(null)
                    setSending(false)
                },
                onError: (errors: ErrorLike) => {
                    setResendEmailError(asError(errors))
                    setResentEmailTo(null)
                    setSending(false)
                },
            })
        }
    }, [authenticatedUser])

    useEffect(() => {
        telemetryService.log('VerifyEmailToastViewed')
    }, [telemetryService])

    return (
        <Modal className={styles.codySurveyToastModal} position="center" aria-label="Welcome message">
            <H3 className="mb-4">
                <Icon svgPath={mdiEmail} className={classNames('mr-2', styles.emailIcon)} aria-hidden={true} />
                Verify your email address
            </H3>
            <Text>To use Cody, our AI Assistant, you'll need to verify your email address.</Text>
            <Text className="d-flex align-items-center">
                <span className="mr-1">Didn't get an email?</span>
                {sending ? (
                    <span>Sending...</span>
                ) : (
                    <>
                        <span>Click to </span>
                        <Button variant="link" className={classNames('p-0 ml-1', styles.resendButton)} onClick={resend}>
                            resend
                        </Button>
                        .
                    </>
                )}
            </Text>
            {resentEmailTo && (
                <Text>
                    Sent verification email to <strong>{resentEmailTo}</strong>.
                </Text>
            )}
            {resendEmailError && <Text>{resendEmailError.message}.</Text>}
            <div className="d-flex justify-content-end mt-4">
                <Button
                    className={styles.codySurveyToastModalButton}
                    variant="primary"
                    onClick={onNext}
                    disabled={!hasVerifiedEmail}
                >
                    Next
                </Button>
            </div>
        </Modal>
    )
}

export const CodySurveyToast: React.FC<
    {
        authenticatedUser: AuthenticatedUser
    } & TelemetryProps
> = ({ authenticatedUser, telemetryService }) => {
    const [showVerifyEmail, setShowVerifyEmail] = useState(isEmailVerificationNeededForCody())
    const [updatePostSignupCompletion] = useMutation<setCompletedPostSignupResult, setCompletedPostSignupVariables>(
        SET_COMPLETED_POST_SIGNUP,
        {
            variables: {
                userID: authenticatedUser.id,
            },
        }
    )

    const handleSubmitEnd = (): JSX.Element => {
        // eslint-disable-next-line no-console
        updatePostSignupCompletion().catch(console.error)

        return <Navigate to={PageRoutes.GetCody} replace={true} />
    }

    const dismissVerifyEmail = useCallback(() => {
        telemetryService.log('VerifyEmailToastDismissed')
        setShowVerifyEmail(false)
    }, [telemetryService])

    if (authenticatedUser.completedPostSignup) {
        return <Navigate to={PageRoutes.GetCody} replace={true} />
    }

    if (showVerifyEmail) {
        return (
            <CodyVerifyEmailToast
                onNext={dismissVerifyEmail}
                authenticatedUser={authenticatedUser}
                telemetryService={telemetryService}
                hasVerifiedEmail={authenticatedUser.hasVerifiedEmail}
            />
        )
    }

    return <CodySurveyToastInner onSubmitEnd={handleSubmitEnd} telemetryService={telemetryService} />
}
