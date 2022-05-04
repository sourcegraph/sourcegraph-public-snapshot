import React from 'react'

import { Button, Checkbox } from '@sourcegraph/wildcard'

import { SurveyRatingRadio } from './SurveyRatingRadio'
import { Toast } from './Toast'

import styles from './SurveyUserRatingForm.module.scss'

export interface SurveyUserRatingFormProps {
    onChange?: (score: number) => void
    toggleErrorMessage: boolean
    onContinue: () => void
    onDismiss: () => void
    shouldPermanentlyDismiss: boolean
    toggleShouldPermanentlyDismiss: (value: boolean) => void
}

export const SurveyUserRatingForm: React.FunctionComponent<SurveyUserRatingFormProps> = ({
    onChange,
    toggleErrorMessage,
    onDismiss,
    onContinue,
    shouldPermanentlyDismiss,
    toggleShouldPermanentlyDismiss,
}) => (
    <Toast
        title="Tell us what you think"
        subtitle={
            <span id="survey-toast-scores">How likely is it that you would recommend Sourcegraph to a friend?</span>
        }
        cta={
            <>
                <SurveyRatingRadio ariaLabelledby="survey-form-scores" onChange={onChange} />
                {toggleErrorMessage && (
                    <div className={styles.alertDanger} role="alert">
                        Please select a score between 0 to 10
                    </div>
                )}
            </>
        }
        footer={
            <div className="d-flex align-items-center justify-content-between">
                <Checkbox
                    id="survey-toast-refuse"
                    label="Don't show this again"
                    checked={shouldPermanentlyDismiss}
                    onChange={event => toggleShouldPermanentlyDismiss(event.target.checked)}
                />
                <Button variant="secondary" size="sm" onClick={onContinue}>
                    Continue
                </Button>
            </div>
        }
        onDismiss={onDismiss}
    />
)
