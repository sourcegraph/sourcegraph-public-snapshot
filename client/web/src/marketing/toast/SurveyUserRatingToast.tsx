import React from 'react'

import { Button, Checkbox, Link, Text } from '@sourcegraph/wildcard'

import { SurveyRatingRadio } from '../components/SurveyRatingRadio'

import { Toast } from './Toast'

import styles from './SurveyUserRatingToast.module.scss'

export interface SurveyUserRatingToastProps {
    score: number
    onChange: (score: number) => void
    toggleErrorMessage: boolean
    onContinue: () => void
    onDismiss: () => void
    setToggledPermanentlyDismiss: (value: boolean) => void
}

export const SurveyUserRatingToast: React.FunctionComponent<SurveyUserRatingToastProps> = ({
    score,
    onChange,
    toggleErrorMessage,
    onDismiss,
    onContinue,
    setToggledPermanentlyDismiss,
}) => (
    <Toast
        title="Tell us what you think"
        subtitle={
            <span id="survey-toast-scores">How likely is it that you would recommend Sourcegraph to a friend?</span>
        }
        cta={
            <>
                <SurveyRatingRadio ariaLabelledby="survey-toast-scores" score={score} onChange={onChange} />
                {toggleErrorMessage && (
                    <div className={styles.alertDanger} role="alert">
                        Please select a score between 0 to 10
                    </div>
                )}
            </>
        }
        footer={
            <>
                <Text className="d-flex align-items-center justify-content-between mb-1">
                    <span>
                        By submitting your feedback, you agree to the{' '}
                        <Link to="https://sourcegraph.com/terms/privacy">Sourcegraph Privacy Policy</Link>.
                    </span>
                </Text>
                <div className="d-flex align-items-center justify-content-between">
                    <Checkbox
                        id="survey-toast-refuse"
                        label={<span className={styles.checkboxLabel}>Don't show this again</span>}
                        onChange={event => setToggledPermanentlyDismiss(event.target.checked)}
                    />
                    <Button variant="secondary" size="sm" onClick={onContinue}>
                        Continue
                    </Button>
                </div>
            </>
        }
        onDismiss={onDismiss}
    />
)
