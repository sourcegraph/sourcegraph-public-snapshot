import React from 'react'

import type { ApolloError } from '@apollo/client'

import { Alert, Button } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { SurveyUseCaseForm } from '../components/SurveyUseCaseForm'

import type { TotalFeedbackState } from './SurveyToastContent'
import { Toast } from './Toast'

import styles from './SurveyUseCaseToast.module.scss'

export interface UseCaseFeedbackState extends Pick<TotalFeedbackState, 'otherUseCase' | 'better' | 'email'> {}

export interface UseCaseFeedbackModifiers {
    onChangeOtherUseCase: (otherUseCase: TotalFeedbackState['otherUseCase']) => void
    onChangeBetter: (additionalInfo: TotalFeedbackState['better']) => void
    onChangeEmail: (email: TotalFeedbackState['email']) => void
}

interface SurveyUseCaseFormToastProps extends UseCaseFeedbackState, UseCaseFeedbackModifiers {
    isSubmitting?: boolean
    onDismiss: () => void
    onDone: () => Promise<void>
    authenticatedUser: AuthenticatedUser | null
    error?: ApolloError
}

export const SurveyUseCaseToast: React.FunctionComponent<SurveyUseCaseFormToastProps> = ({
    isSubmitting,
    onDismiss,
    onDone,
    otherUseCase,
    onChangeOtherUseCase,
    better,
    onChangeBetter,
    email,
    onChangeEmail,
    authenticatedUser,
    error,
}) => (
    <Toast
        toastBodyClassName={styles.toastBody}
        toastContentClassName="mt-0"
        cta={
            <SurveyUseCaseForm
                authenticatedUser={authenticatedUser}
                otherUseCase={otherUseCase}
                onChangeOtherUseCase={onChangeOtherUseCase}
                better={better}
                onChangeBetter={onChangeBetter}
                email={email}
                onChangeEmail={onChangeEmail}
            />
        }
        footer={
            <>
                {error && (
                    <div className="d-flex">
                        <Alert variant="danger">Error: {error.message}</Alert>
                    </div>
                )}
                <div className="d-flex justify-content-end">
                    <Button variant="primary" size="sm" onClick={onDone} disabled={isSubmitting}>
                        Done
                    </Button>
                </div>
            </>
        }
        onDismiss={onDismiss}
    />
)
