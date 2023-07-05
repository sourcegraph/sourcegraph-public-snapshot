import { useState, useCallback, useEffect } from 'react'

import { mdiEmail } from '@mdi/js'
import classNames from 'classnames'

import { asError, ErrorLike } from '@sourcegraph/common'
import { gql, useMutation } from '@sourcegraph/http-client'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Checkbox, Form, H3, Modal, Text, Button, Icon, useCookieStorage } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { CodyColorIcon } from '../../cody/chat/CodyPageIcon'
import { isEmailVerificationNeededForCody } from '../../cody/isCodyEnabled'
import { LoaderButton } from '../../components/LoaderButton'
import { SubmitCodySurveyResult, SubmitCodySurveyVariables } from '../../graphql-operations'
import { resendVerificationEmail } from '../../user/settings/emails/UserEmail'

import styles from './CodySurveyToast.module.scss'

const SUBMIT_CODY_SURVEY = gql`
    mutation SubmitCodySurvey($isForWork: Boolean!, $isForPersonal: Boolean!) {
        submitCodySurvey(isForWork: $isForWork, isForPersonal: $isForPersonal) {
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
                    />
                </div>
            </Form>
        </Modal>
    )
}

const CodyVerifyEmailToast: React.FC<{ onNext: () => void; authenticatedUser: AuthenticatedUser } & TelemetryProps> = ({
    onNext,
    authenticatedUser,
    telemetryService,
}) => {
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
                <Button className={styles.codySurveyToastModalButton} variant="primary" onClick={onNext}>
                    Next
                </Button>
            </div>
        </Modal>
    )
}

export const useCodySurveyToast = (): {
    show: boolean
    dismiss: () => void
    setShouldShowCodySurvey: (show: boolean) => void
} => {
    // we specifically use cookie storage as we want consistent value between when user is logged out and logged in / signed up
    // as well as cross-domain such about.sourcegraph.com
    const [shouldShowCodySurvey, setShouldShowCodySurvey] = useCookieStorage<boolean>('cody.survey.show', false, {
        expires: 365,
    })
    const [hasSubmitted, setHasSubmitted] = useTemporarySetting('cody.survey.submitted', false)
    const dismiss = useCallback(() => {
        setHasSubmitted(true)
        setShouldShowCodySurvey(false)
    }, [setHasSubmitted, setShouldShowCodySurvey])

    useEffect(() => {
        if (shouldShowCodySurvey && hasSubmitted) {
            setShouldShowCodySurvey(false)
        }
    }, [shouldShowCodySurvey, hasSubmitted, setShouldShowCodySurvey])

    return {
        // we calculate "show" value based whether this a new signup and whether they already have submitted survey
        show: !hasSubmitted && !!shouldShowCodySurvey,
        dismiss,
        setShouldShowCodySurvey,
    }
}

export const CodySurveyToast: React.FC<
    {
        authenticatedUser?: AuthenticatedUser
    } & TelemetryProps
> = ({ authenticatedUser, telemetryService }) => {
    const { show, dismiss } = useCodySurveyToast()
    const [showVerifyEmail, setShowVerifyEmail] = useState(show && isEmailVerificationNeededForCody())
    const dismissVerifyEmail = useCallback(() => {
        telemetryService.log('VerifyEmailToastDismissed')
        setShowVerifyEmail(false)
    }, [telemetryService])

    if (!show) {
        return null
    }

    if (showVerifyEmail && authenticatedUser) {
        return (
            <CodyVerifyEmailToast
                onNext={dismissVerifyEmail}
                authenticatedUser={authenticatedUser}
                telemetryService={telemetryService}
            />
        )
    }

    return <CodySurveyToastInner onSubmitEnd={dismiss} telemetryService={telemetryService} />
}
