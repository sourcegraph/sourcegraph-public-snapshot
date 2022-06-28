import React from 'react'

import { Button, FeedbackText, H4 } from '@sourcegraph/wildcard'

import { Toast } from './Toast'

import styles from './SurveySuccessToast.module.scss'

interface SurveySuccessToastProps {
    onDismiss: () => void
}

export const SurveySuccessToast: React.FunctionComponent<SurveySuccessToastProps> = ({ onDismiss }) => (
    <Toast
        subtitle={<H4 className={styles.toastSubtitle}>Thank you for your feedback!</H4>}
        cta={<FeedbackText headerText="Anything else?" />}
        footer={
            <div className="d-flex justify-content-end">
                <Button variant="primary" size="sm" onClick={onDismiss}>
                    Done
                </Button>
            </div>
        }
        className="text-center"
        onDismiss={onDismiss}
    />
)
