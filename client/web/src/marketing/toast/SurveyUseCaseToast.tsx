import React from 'react'

import { Button } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { SurveyUseCaseForm } from '../components/SurveyUseCaseForm'

import { TotalFeedbackState } from './SurveyToastContent'
import { Toast } from './Toast'

import styles from './SurveyUseCaseToast.module.scss'

export interface UseCaseFeedbackState
    extends Pick<TotalFeedbackState, 'useCases' | 'otherUseCase' | 'additionalInformation' | 'email'> {}

export interface UseCaseFeedbackModifiers {
    onChangeUseCase: (useCase: TotalFeedbackState['useCases'][number]) => void
    onChangeOtherUseCase: (otherUseCase: TotalFeedbackState['otherUseCase']) => void
    onChangeAdditionalInformation: (additionalInfo: TotalFeedbackState['additionalInformation']) => void
    onChangeEmail: (email: TotalFeedbackState['email']) => void
}

interface SurveyUseCaseFormToastProps extends UseCaseFeedbackState, UseCaseFeedbackModifiers {
    isSubmitting?: boolean
    onDismiss: () => void
    onDone: () => Promise<void>
    authenticatedUser: AuthenticatedUser | null
}

export const SurveyUseCaseToast: React.FunctionComponent<SurveyUseCaseFormToastProps> = ({
    isSubmitting,
    onDismiss,
    onDone,
    useCases,
    onChangeUseCase,
    otherUseCase,
    onChangeOtherUseCase,
    additionalInformation,
    onChangeAdditionalInformation,
    email,
    onChangeEmail,
    authenticatedUser,
}) => (
    <Toast
        toastBodyClassName={styles.toastBody}
        toastContentClassName="mt-0"
        cta={
            <SurveyUseCaseForm
                title="You are using sourcegraph to..."
                authenticatedUser={authenticatedUser}
                useCases={useCases}
                onChangeUseCase={onChangeUseCase}
                otherUseCase={otherUseCase}
                onChangeOtherUseCase={onChangeOtherUseCase}
                additionalInformation={additionalInformation}
                onChangeAdditionalInformation={onChangeAdditionalInformation}
                email={email}
                onChangeEmail={onChangeEmail}
            />
        }
        footer={
            <div className="d-flex justify-content-end">
                <Button variant="primary" size="sm" onClick={onDone} disabled={isSubmitting}>
                    Done
                </Button>
            </div>
        }
        onDismiss={onDismiss}
    />
)
