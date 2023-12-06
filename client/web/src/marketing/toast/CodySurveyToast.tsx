import { useState, useCallback, useEffect } from 'react'

import { mdiEmail } from '@mdi/js'
import classNames from 'classnames'
import { useLocation } from 'react-router-dom'

import { asError, type ErrorLike } from '@sourcegraph/common'
import { gql, useMutation } from '@sourcegraph/http-client'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Checkbox, Form, H3, Modal, Text, Button, Icon, AnchorLink } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { getReturnTo } from '../../auth/SignInSignUpCommon'
import { CodyColorIcon } from '../../cody/chat/CodyPageIcon'
import { LoaderButton } from '../../components/LoaderButton'
import type {
    SubmitCodySurveyResult,
    SubmitCodySurveyVariables,
    SetCompletedPostSignupVariables,
    SetCompletedPostSignupResult,
} from '../../graphql-operations'
import { PageRoutes } from '../../routes.constants'
import { resendVerificationEmail } from '../../user/settings/emails/UserEmail'

import styles from './CodySurveyToast.module.scss'

export const SUBMIT_CODY_SURVEY = gql`
    mutation SubmitCodySurvey($isForWork: Boolean!, $isForPersonal: Boolean!) {
        submitCodySurvey(isForWork: $isForWork, isForPersonal: $isForPersonal) {
            alwaysNil
        }
    }
`

const SET_COMPLETED_POST_SIGNUP = gql`
    mutation SetCompletedPostSignup($userID: ID!) {
        setCompletedPostSignup(userID: $userID) {
            alwaysNil
        }
    }
`

const CodySurveyToastInner: React.FC<
    { onSubmitEnd: () => void; userId: string; hasVerifiedEmail: boolean } & TelemetryProps & TelemetryV2Props
> = ({ userId, onSubmitEnd, telemetryService, telemetryRecorder, hasVerifiedEmail }) => {
    const [isCodyForWork, setIsCodyForWork] = useState(false)
    const [isCodyForPersonalStuff, setIsCodyForPersonalStuff] = useState(false)

    const handleCodyForWorkChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setIsCodyForWork(event.target.checked)
    }, [])
    const handleCodyForPersonalStuffChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setIsCodyForPersonalStuff(event.target.checked)
    }, [])

    const [submitCodySurvey, { loading: loadingCodySurvey, error: submitSurveyError }] = useMutation<
        SubmitCodySurveyResult,
        SubmitCodySurveyVariables
    >(SUBMIT_CODY_SURVEY, {
        variables: {
            isForWork: isCodyForWork,
            isForPersonal: isCodyForPersonalStuff,
        },
    })

    const [updatePostSignupCompletion, { loading: loadingPostSignup, error: setPostSignupError }] = useMutation<
        SetCompletedPostSignupResult,
        SetCompletedPostSignupVariables
    >(SET_COMPLETED_POST_SIGNUP, {
        variables: {
            userID: userId,
        },
    })

    const loading = loadingCodySurvey || loadingPostSignup
    const error = !!submitSurveyError || !!setPostSignupError

    const handleSubmit = useCallback(
        async (event: React.FormEvent<HTMLFormElement>) => {
            const eventParams = { isCodyForPersonalStuff, isCodyForWork }
            telemetryService.log('CodyUsageToastSubmitted', eventParams, eventParams)
            telemetryRecorder.recordEvent('CodyUsageToast', 'submitted', {
                privateMetadata: { eventParams },
            })
            event.preventDefault()

            try {
                await submitCodySurvey()

                if (hasVerifiedEmail) {
                    await updatePostSignupCompletion()
                }

                onSubmitEnd()
            } catch (error) {
                /* eslint-disable no-console */
                console.error(error)
            }
        },
        [
            hasVerifiedEmail,
            isCodyForPersonalStuff,
            isCodyForWork,
            onSubmitEnd,
            submitCodySurvey,
            updatePostSignupCompletion,
            telemetryService,
            telemetryRecorder,
        ]
    )

    useEffect(() => {
        telemetryService.log('CodySurveyToastViewed')
        telemetryRecorder.recordEvent('CodySurveyToast', 'viewed')
    }, [telemetryService, telemetryRecorder])

    return (
        <Modal
            className={styles.codySurveyToastModal}
            position="center"
            aria-label="Welcome message"
            containerClassName={styles.modalOverlay}
        >
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
                {error && (
                    <Text size="small" className="text-danger mt-3 mb-2">
                        An error occurred. Please reload the page and try again. If this persists, contact support at
                        support@sourcegraph.com
                    </Text>
                )}
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
    { onNext: () => void; authenticatedUser: AuthenticatedUser } & TelemetryProps & TelemetryV2Props
> = ({ onNext, authenticatedUser, telemetryService, telemetryRecorder }) => {
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
        telemetryRecorder.recordEvent('VerifyEmailToast', 'viewed')
    }, [telemetryService, telemetryRecorder])

    return (
        <Modal
            className={styles.codySurveyToastModal}
            position="center"
            aria-label="Welcome message"
            containerClassName={styles.modalOverlay}
        >
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
                <AnchorLink className="mr-3 mt-auto mb-auto" to="/-/sign-out">
                    Sign out
                </AnchorLink>
                <Button className={styles.codySurveyToastModalButton} variant="primary" onClick={onNext}>
                    Next
                </Button>
            </div>
        </Modal>
    )
}

export const CodySurveyToast: React.FC<
    {
        authenticatedUser: AuthenticatedUser
    } & TelemetryProps &
        TelemetryV2Props
> = ({ authenticatedUser, telemetryService, telemetryRecorder }) => {
    const [showVerifyEmail, setShowVerifyEmail] = useState(!authenticatedUser.hasVerifiedEmail)

    const location = useLocation()

    const handleSubmitEnd = (): void => {
        // Redirects once user submits the post-sign-up form
        const returnTo = getReturnTo(location, PageRoutes.GetCody)
        window.location.replace(returnTo)
    }

    const dismissVerifyEmail = useCallback(() => {
        telemetryService.log('VerifyEmailToastDismissed')
        telemetryRecorder.recordEvent('VerifyEmailToast', 'dismissed')
        setShowVerifyEmail(false)
    }, [telemetryService, telemetryRecorder])

    if (showVerifyEmail) {
        return (
            <CodyVerifyEmailToast
                onNext={dismissVerifyEmail}
                authenticatedUser={authenticatedUser}
                telemetryService={telemetryService}
                telemetryRecorder={telemetryRecorder}
            />
        )
    }

    return (
        <CodySurveyToastInner
            telemetryService={telemetryService}
            telemetryRecorder={telemetryRecorder}
            onSubmitEnd={handleSubmitEnd}
            userId={authenticatedUser.id}
            hasVerifiedEmail={authenticatedUser.hasVerifiedEmail}
        />
    )
}
