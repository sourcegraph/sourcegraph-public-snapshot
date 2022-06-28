import React from 'react'

import { Button } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { SurveyUseCaseForm } from '../components/SurveyUseCaseForm'

import { TotalFeedbackState } from './SurveyToastContent'
import { Toast } from './Toast'

import styles from './SurveyUseCaseToast.module.scss'

export interface UseCaseFeedbackState
    extends Pick<TotalFeedbackState, 'useCases' | 'otherUseCase' | 'better' | 'email'> {}

export interface UseCaseFeedbackModifiers {
    onChangeUseCases: (useCases: TotalFeedbackState['useCases']) => void
    onChangeOtherUseCase: (otherUseCase: TotalFeedbackState['otherUseCase']) => void
    onChangeBetter: (additionalInfo: TotalFeedbackState['better']) => void
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
    onChangeUseCases,
    otherUseCase,
    onChangeOtherUseCase,
    better,
    onChangeBetter,
    email,
    onChangeEmail,
    authenticatedUser,
}) => (
    <Toast
        toastBodyClassName={styles.toastBody}
        toastContentClassName="mt-0"
        cta={
            <SurveyUseCaseForm
                title="You are using Sourcegraph to..."
                authenticatedUser={authenticatedUser}
                useCases={useCases}
                onChangeUseCases={onChangeUseCases}
                otherUseCase={otherUseCase}
                onChangeOtherUseCase={onChangeOtherUseCase}
                better={better}
                onChangeBetter={onChangeBetter}
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
