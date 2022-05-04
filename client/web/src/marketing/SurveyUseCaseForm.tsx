import React, { useState } from 'react'

import { Button, FlexTextArea, TextArea } from '@sourcegraph/wildcard'

import { SurveyUseCaseCheckbox } from './SurveyUseCaseCheckbox'
import { Toast } from './Toast'

import styles from './SurveyUseCaseForm.module.scss'

export const OPTIONS = [
    {
        id: 'understandCodebase',
        labelValue: 'Understand a new codebase',
    },
    {
        id: 'fixSecurityVulnerability',
        labelValue: 'Fix security vulnerabilities',
    },
    {
        id: 'reuseCode',
        labelValue: 'Reuse code',
    },
    {
        id: 'respondToIncidents',
        labelValue: 'Respond to incidents',
    },
    {
        id: 'ImproveCodeQuality',
        labelValue: 'Improve code quality',
    },
    {
        id: 'other',
        labelValue: 'Other',
    },
]

interface FormStateType {
    moreSharedInfo: string
    otherUseCase: string
    useCases: string[]
}

interface SurveyUseCaseFormProps {
    onDismiss: () => void
    handleDone: (props: FormStateType) => void
}

export const SurveyUseCaseForm: React.FunctionComponent<SurveyUseCaseFormProps> = ({ onDismiss, handleDone }) => {
    const [useCases, setUseCases] = useState<string[]>([])
    const [otherUseCase, setOtherUseCase] = useState<string>('')
    const [moreSharedInfo, setMoreSharedInfo] = useState<string>('')

    const handleSubmit = (): void => {
        handleDone({
            useCases,
            otherUseCase,
            moreSharedInfo,
        })
    }

    const handleSelectUseCase = (value: string): void => {
        if (useCases.includes(value)) {
            setUseCases(current => current.filter(instance => instance !== value))
            return
        }
        setUseCases(current => [...current, value])
    }

    return (
        <Toast
            toastBodyClassName={styles.toastBody}
            subtitle={
                <span id="usecase-group" className={styles.toastSubtitle}>
                    You are using sourcegraph to...
                </span>
            }
            cta={
                <div className="mb-2">
                    <div className={styles.checkWrap}>
                        {OPTIONS.map(({ id, labelValue }) => (
                            <SurveyUseCaseCheckbox
                                onChange={() => handleSelectUseCase(id)}
                                key={id}
                                id={id}
                                checked={useCases.includes(id)}
                                label={labelValue}
                                aria-labelledby={`usecase-group ${id}`}
                            />
                        ))}
                    </div>
                    {useCases.includes('other') && (
                        <FlexTextArea
                            containerClassName="mt-3"
                            label={
                                <span className={styles.textareaLabel}>What else are you using sourcegraph to do?</span>
                            }
                            name="other"
                            placeholder="Find..."
                            onChange={event => setOtherUseCase(event.target.value)}
                            value={otherUseCase}
                        />
                    )}
                    <TextArea
                        className="mt-3"
                        size="small"
                        name="more"
                        onChange={event => setMoreSharedInfo(event.target.value)}
                        value={moreSharedInfo}
                        label={
                            <span className={styles.textareaLabel}>Anything else you would like to share with us?</span>
                        }
                    />
                </div>
            }
            footer={
                <div className="d-flex justify-content-end">
                    <Button variant="primary" size="sm" onClick={handleSubmit}>
                        Done
                    </Button>
                </div>
            }
            onDismiss={onDismiss}
        />
    )
}
