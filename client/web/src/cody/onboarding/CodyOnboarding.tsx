import React, { useEffect, useState } from 'react'

import { useNavigate } from 'react-router-dom'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import type { TelemetryRecorder, TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Modal, useSearchParameters } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'

import { EditorStep } from './EditorStep'
import { PurposeStep } from './PurposeStep'
import { WelcomeStep } from './WelcomeStep'

import styles from './CodyOnboarding.module.scss'

export interface IEditor {
    id: number // a unique number identifier for telemetry
    icon: string
    name: string
    publisher: string
    releaseStage: string
    docs?: string
    instructions?: React.FC<{
        onBack?: () => void
        onClose: () => void
        showStep?: number
        telemetryRecorder: TelemetryRecorder
    }>
}

interface CodyOnboardingProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser | null
}

export function CodyOnboarding({ authenticatedUser, telemetryRecorder }: CodyOnboardingProps): JSX.Element | null {
    const [showEditorStep, setShowEditorStep] = useState(false)
    const [completed = false, setOnboardingCompleted] = useTemporarySetting('cody.onboarding.completed', false)
    const [signUpFlowEnabled, signUpFlowStatus] = useFeatureFlag('ab-shortened-install-first-signup-flow-cody-2024-04')
    // steps start from 0
    const [step = -1, setOnboardingStep] = useTemporarySetting('cody.onboarding.step', 0)

    const onNext = (): void => setOnboardingStep(currentsStep => (currentsStep || 0) + 1)

    const parameters = useSearchParameters()
    const enrollPro = parameters.get('pro') === 'true'
    const returnToURL = parameters.get('returnTo')
    const isCody = parameters.get('requestFrom') === 'CODY'

    const navigate = useNavigate()

    useEffect(() => {
        if (completed && returnToURL) {
            navigate(returnToURL)
        }
    }, [completed, returnToURL, navigate])

    useEffect(() => {
        if (signUpFlowStatus === 'loaded' && signUpFlowEnabled && isCody) {
            setOnboardingStep(currentsStep => (currentsStep || 0) + 2)
            setOnboardingCompleted(true)
            setShowEditorStep(true)
        }
        if (signUpFlowStatus === 'loaded' && isCody) {
            const metadataKey = signUpFlowEnabled ? 'treatmentVariant' : 'controlVariant'
            telemetryRecorder.recordEvent('cody.onboarding.ABShortenedSignupFlowForInstalls202404', 'enroll', {
                metadata: { [metadataKey]: 1 },
            })
        }
    }, [signUpFlowEnabled, signUpFlowStatus, isCody, setOnboardingStep, setOnboardingCompleted, telemetryRecorder])

    if (completed && returnToURL) {
        return null
    }

    if (!showEditorStep && (completed || step === -1 || step > 1)) {
        return null
    }

    if (!authenticatedUser) {
        return null
    }

    if (signUpFlowStatus !== 'loaded') {
        return null
    }

    const handleShowLastStep = (): void => {
        setOnboardingCompleted(true)
        setShowEditorStep(true)
        telemetryRecorder.recordEvent('cody.onboarding.hubspotForm.fromWorkPersonalToHandRaiserTest', 'enroll', {
            metadata: { controlVariant: 1 },
        })
    }

    return (
        <Modal
            isOpen={true}
            position="center"
            aria-label="Cody Onboarding"
            className={styles.modal}
            containerClassName={styles.root}
        >
            {step === 0 && <WelcomeStep onNext={onNext} pro={enrollPro} telemetryRecorder={telemetryRecorder} />}
            {step === 1 && (
                <PurposeStep
                    authenticatedUser={authenticatedUser}
                    onNext={() => {
                        onNext()
                        handleShowLastStep()
                    }}
                    pro={enrollPro}
                    telemetryRecorder={telemetryRecorder}
                />
            )}
            {showEditorStep && (
                <EditorStep
                    onCompleted={() => {
                        setShowEditorStep(false)
                    }}
                    pro={enrollPro}
                    telemetryRecorder={telemetryRecorder}
                />
            )}
        </Modal>
    )
}
