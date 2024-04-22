import { useEffect } from 'react'

import type { TelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { H2, Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { HubSpotForm } from '../../marketing/components/HubSpotForm'

export function PurposeStep({
    onNext,
    pro,
    authenticatedUser,
    telemetryRecorder,
}: {
    onNext: () => void
    pro: boolean
    authenticatedUser: AuthenticatedUser
    telemetryRecorder: TelemetryRecorder
}): JSX.Element {
    useEffect(() => {
        telemetryRecorder.recordEvent('cody.onboarding.purpose', 'view', { metadata: { tier: pro ? 1 : 0 } })
    }, [pro, telemetryRecorder])

    const primaryEmail = authenticatedUser.emails.find(email => email.isPrimary)?.email

    const handleFormSubmit = (form: HTMLFormElement): void => {
        const choice = form[0].querySelector<HTMLInputElement>('input[name="cody_form_hand_raiser"]')

        if (choice) {
            telemetryRecorder.recordEvent('cody.onboarding.purpose', 'select', {
                metadata: { onboardingCall: choice.checked ? 1 : 0 },
            })
        }
    }

    return (
        <>
            <div className="border-bottom pb-3 mb-3">
                <H2 className="mb-1">Would you like to learn more about Cody for enterprise?</H2>
                <Text className="mb-0 text-muted" size="small">
                    If you're not ready for a conversation, we'll stick to sharing Cody onboarding resources.
                </Text>
            </div>
            <div className="d-flex align-items-center border-bottom mb-3 pb-3 justify-content-center">
                <HubSpotForm
                    formId="19f34edd-1a98-4fc9-9b2b-c1edca727720"
                    onFormSubmitted={() => {
                        onNext()
                    }}
                    onFormLoadError={() => {
                        onNext()
                    }}
                    userId={authenticatedUser.id}
                    userEmail={primaryEmail}
                    masterFormName="qualificationSurvey"
                    onFormSubmit={handleFormSubmit}
                />
            </div>
        </>
    )
}
